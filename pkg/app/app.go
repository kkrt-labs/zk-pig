package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/MadAppGang/httplog"
	httplogzap "github.com/MadAppGang/httplog/zap"
	"github.com/hellofresh/health-go/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	kkrtnet "github.com/kkrt-labs/go-utils/net"
	kkrthttp "github.com/kkrt-labs/go-utils/net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Main         ServerConfig `mapstructure:"main"`
	Healthz      ServerConfig `mapstructure:"healthz"`
	StartTimeout string       `mapstructure:"start-timeout"`
	StopTimeout  string       `mapstructure:"stop-timeout"`
}

type ServerConfig struct {
	Entrypoint        EntrypointConfig `mapstructure:"entrypoint"`
	ReadTimeout       string           `mapstructure:"read-timeout"`
	ReadHeaderTimeout string           `mapstructure:"read-header-timeout"`
	WriteTimeout      string           `mapstructure:"write-timeout"`
	IdleTimeout       string           `mapstructure:"idle-timeout"`
}

type EntrypointConfig struct {
	Network   string `mapstructure:"network"`
	Address   string `mapstructure:"address"`
	KeepAlive string `mapstructure:"keep-alive"`
}

func newServer(cfg *ServerConfig) (*http.Server, error) {
	readTimeout, err := time.ParseDuration(cfg.ReadTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse read timeout: %w", err)
	}
	readHeaderTimeout, err := time.ParseDuration(cfg.ReadHeaderTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse read header timeout: %w", err)
	}
	writeTimeout, err := time.ParseDuration(cfg.WriteTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse write timeout: %w", err)
	}
	idleTimeout, err := time.ParseDuration(cfg.IdleTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse idle timeout: %w", err)
	}

	return &http.Server{
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}, nil
}

type App struct {
	cfg *Config

	name    string
	version string

	services map[string]*Service

	top     *Service
	current *Service

	done chan os.Signal

	logger *zap.Logger

	main       *kkrthttp.Server
	mainRouter *httprouter.Router

	healthz       *kkrthttp.Server
	healthzRouter *httprouter.Router

	liveHealth  *health.Health
	readyHealth *health.Health

	prometheus *prometheus.Registry
}

func NewApp(cfg *Config, opts ...Option) (*App, error) {
	app := &App{
		cfg:           cfg,
		services:      make(map[string]*Service),
		done:          make(chan os.Signal),
		logger:        zap.NewNop(),
		mainRouter:    httprouter.New(),
		healthzRouter: httprouter.New(),
		prometheus:    prometheus.NewRegistry(),
	}

	for _, opt := range opts {
		if err := opt(app); err != nil {
			return nil, err
		}
	}

	app.liveHealth = newHealth(app)
	app.readyHealth = newHealth(app)

	app.registerBaseMetrics()

	return app, nil
}

func newHealth(app *App) *health.Health {
	h, _ := health.New(health.WithComponent(health.Component{Name: app.name, Version: app.version}))
	return h
}

func (app *App) Provide(name string, constructor func() (any, error), opts ...ServiceOption) any {
	if name == "" {
		name = reflect.TypeOf(constructor).Out(0).String()
	}

	if svc, ok := app.services[name]; ok {
		app.current.addDep(svc) // current can not be nil here
		return svc.value
	}

	svc := app.createService(name, constructor, opts...)
	app.services[name] = svc

	return svc.value
}

func (app *App) createService(name string, constructor func() (any, error), opts ...ServiceOption) *Service {
	previous := app.current
	svc := newService(name, constructor, opts...)
	svc.app = app

	app.current = svc // set the current service pointer
	svc.construct()   // construct can perform calls to Provide moving the current service pointer
	if previous != nil {
		previous.addDep(svc)
	} else {
		app.top = svc
	}

	app.current = previous // restore the current service pointer

	return svc
}

func Provide[T any](app *App, name string, constructor func() (T, error), opts ...ServiceOption) T {
	val := app.Provide(name, func() (any, error) {
		return constructor()
	}, opts...)

	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Invalid {
		// return zero value for the type T
		var zero T
		return zero
	}

	return val.(T)
}

func (app *App) Error() error {
	if app.top == nil || app.top.err == nil {
		return nil
	}
	return app.top.err
}

func (app *App) Start(ctx context.Context) error {
	app.logger.Info("Starting app...")
	err := app.start(ctx)
	if err != nil {
		app.logger.Error("Failed to start app", zap.Error(err))
		return err
	}
	app.logger.Info("App started")
	return nil
}

func (app *App) start(ctx context.Context) error {
	if app.top == nil {
		return fmt.Errorf("no service constructed yet")
	}

	if app.top.err != nil {
		return app.top.err
	}

	app.setHandlers()

	if err := app.top.start(ctx); err != nil {
		return err
	}

	return nil
}

func (app *App) Stop(ctx context.Context) error {
	app.logger.Info("Stopping app...")
	err := app.stop(ctx)
	if err != nil {
		app.logger.Error("Failed to stop app", zap.Error(err))
		return err
	}
	app.logger.Info("App stopped")
	return nil
}

func (app *App) stop(ctx context.Context) error {
	if app.top == nil {
		return fmt.Errorf("no service constructed yet")
	}

	if err := app.top.stop(ctx); err != nil {
		return err
	}

	return nil
}

func (app *App) Run(ctx context.Context) error {
	err := app.Start(ctx)
	if err != nil {
		return err
	}

	app.listenSignals()

	<-app.done

	app.stopListeningSignals()

	return app.Stop(ctx)
}

func (app *App) listenSignals() {
	signal.Notify(app.done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
}

func (app *App) stopListeningSignals() {
	signal.Stop(app.done)
}

func (app *App) EnableMain() {
	app.main = app.server("main", &app.cfg.Main)
}

func (app *App) EnableHealthz() {
	app.healthz = app.server("healthz", &app.cfg.Healthz)
}

func (app *App) server(name string, cfg *ServerConfig) *kkrthttp.Server {
	return Provide(app, fmt.Sprintf("app.%v", name), func() (*kkrthttp.Server, error) {
		return &kkrthttp.Server{
			Entrypoint: app.entrypoint(name, &cfg.Entrypoint),
			Server:     app.httpServer(name, cfg),
		}, nil
	})
}

func (app *App) httpServer(name string, cfg *ServerConfig) *http.Server {
	return Provide(app, fmt.Sprintf("app.%v.server", name), func() (*http.Server, error) {
		return newServer(cfg)
	})
}

func (app *App) entrypoint(name string, cfg *EntrypointConfig) *kkrtnet.Entrypoint {
	return Provide(app, fmt.Sprintf("app.%v.entrypoint", name), func() (*kkrtnet.Entrypoint, error) {
		keepAlive, err := time.ParseDuration(cfg.KeepAlive)
		if err != nil {
			return nil, fmt.Errorf("failed to parse keep alive: %w", err)
		}

		return kkrtnet.NewEntrypoint(&kkrtnet.EntrypointConfig{
			Network:   "tcp",
			Address:   cfg.Address,
			KeepAlive: keepAlive,
		}), nil
	})
}

func (app *App) setHandlers() {
	app.setMainHandler()
	app.setHealthzHandler()
}

func (app *App) setMainHandler() {
	if app.main != nil {
		h := app.instrumentMiddleware().Then(app.mainRouter)
		app.main.Server.Handler = h
	}
}

func (app *App) instrumentMiddleware() alice.Chain {
	return alice.New(
		// Log Requests on main router
		httplog.LoggerWithConfig(httplog.LoggerConfig{
			Formatter: httplogzap.ZapLogger(app.logger, zapcore.InfoLevel, ""),
		}),
		// Instrument main router with prometheus metrics
		func(next http.Handler) http.Handler {
			return promhttp.InstrumentMetricHandler(app.prometheus, next)
		},
	)
}

func (app *App) setHealthzHandler() {
	app.healthzRouter.Handler(http.MethodGet, "/live", app.liveHealth.Handler())
	app.healthzRouter.Handler(http.MethodGet, "/ready", app.readyHealth.Handler())
	app.healthzRouter.Handler(http.MethodGet, "/metrics", promhttp.HandlerFor(app.prometheus, promhttp.HandlerOpts{}))

	if app.healthz != nil {
		app.healthz.Server.Handler = app.healthzRouter
	}
}

func (app *App) registerBaseMetrics() {
	app.prometheus.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	app.prometheus.MustRegister(collectors.NewGoCollector())
}

type ServiceStatus uint32

const (
	Constructing ServiceStatus = iota
	Constructed
	Starting
	Running
	Stopping
	Stopped
	Error
)

type Service struct {
	app *App

	name string

	value any

	constructor func() (any, error)

	deps   map[string]*Service
	depsOf map[string]*Service

	mux    sync.RWMutex
	status atomic.Uint32
	err    *ServiceError

	startOnce sync.Once

	stopOnce sync.Once
	stopChan chan struct{}

	healthConfig *health.Config

	metricsPrefix string
}

func newService(name string, constructor func() (any, error), opts ...ServiceOption) *Service {
	s := &Service{
		name:         name,
		constructor:  constructor,
		deps:         make(map[string]*Service),
		depsOf:       make(map[string]*Service),
		stopChan:     make(chan struct{}),
		healthConfig: new(health.Config),
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			_ = s.fail(err)
			return nil
		}
	}

	return s
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Status() ServiceStatus {
	return ServiceStatus(s.status.Load())
}

func (s *Service) setStatus(status ServiceStatus) {
	s.status.Store(uint32(status))
}

func (s *Service) setStatusWithLock(status ServiceStatus) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.setStatus(status)
}

func (s *Service) fail(err error) *ServiceError {
	if svcErr, ok := err.(*ServiceError); ok {
		s.err = svcErr
	} else {
		s.err = &ServiceError{
			service:   s,
			directErr: err,
		}
	}
	s.setStatus(Error)

	return s.err
}

func (s *Service) failWithLock(err error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	_ = s.fail(err)
}

func (s *Service) construct() {
	s.setStatus(Constructing)
	val, constructorErr := s.constructor()
	if constructorErr != nil {
		_ = s.fail(constructorErr)
		return
	}

	s.value = val
	if err := s.registerReadyCheck(); err != nil {
		_ = s.fail(err)
		return
	}

	s.registerMetric()

	s.setStatus(Constructed)
}

func (s *Service) addDep(dep *Service) {
	if s.isCircularDependency(dep) {
		_ = s.fail(fmt.Errorf("circular dependency detected: %v -> %v", s.name, dep.name))
		return
	}

	// detect circular dependencies
	s.deps[dep.name] = dep
	dep.depsOf[s.name] = s
	if dep.err != nil {
		if s.err == nil {
			_ = s.fail(nil)
		}
		s.err.addDepsErr(dep.err)
	}
}

func (s *Service) isCircularDependency(dep *Service) bool {
	if dep.name == s.name {
		return true
	}
	for _, d := range dep.deps {
		if s.isCircularDependency(d) {
			return true
		}
	}
	return false
}

func (s *Service) start(ctx context.Context) *ServiceError {
	s.startOnce.Do(func() {
		if s.err != nil {
			return
		}

		s.setStatusWithLock(Starting)

		// Start dependencies
		startErr := new(ServiceError)
		wg := sync.WaitGroup{}
		wg.Add(len(s.deps))
		for _, dep := range s.deps {
			go func(dep *Service) {
				defer wg.Done()
				if err := dep.start(ctx); err != nil {
					startErr.addDepsErr(err)
				}
			}(dep)
		}
		wg.Wait()

		if len(startErr.depsErrs) > 0 {
			s.failWithLock(startErr)
			return
		}

		// If all dependencies started successfully then start the service
		if s.err == nil {
			if start, ok := s.value.(Runnable); ok {
				s.app.logger.Info("Starting service", zap.String("service", s.name))
				err := start.Start(ctx)
				if err != nil {
					s.failWithLock(err)
					s.app.logger.Error("Failed to start service", zap.String("service", s.name), zap.Error(err))
					return
				}
				s.app.logger.Info("Service started", zap.String("service", s.name))
			}
		}

		s.setStatusWithLock(Running)
	})

	return s.err
}

func (s *Service) stop(ctx context.Context) *ServiceError {
	if s.err != nil {
		return s.err
	}

	// if one of the dependencies is not running then don't stop
	for _, dep := range s.depsOf {
		if dep.Status() <= Stopping {
			<-s.stopChan
			return s.err
		}
	}

	s.stopOnce.Do(func() {
		if s.err != nil {
			return
		}

		s.setStatusWithLock(Stopping)
		defer func() {
			close(s.stopChan)
		}()

		if stop, ok := s.value.(Runnable); ok {
			s.app.logger.Info("Stopping service", zap.String("service", s.name))
			err := stop.Stop(ctx)
			if err != nil {
				s.failWithLock(err)
				s.app.logger.Error("Failed to stop service", zap.String("service", s.name), zap.Error(err))
				return
			}
			s.app.logger.Info("Service stopped", zap.String("service", s.name))
		}
		if s.err == nil {
			s.setStatusWithLock(Stopped)
		}

		stopErr := new(ServiceError)
		wg := sync.WaitGroup{}
		wg.Add(len(s.deps))
		for _, dep := range s.deps {
			go func(dep *Service) {
				defer wg.Done()
				if err := dep.stop(ctx); err != nil {
					stopErr.addDepsErr(err)
				}
			}(dep)
		}
		wg.Wait()

		if len(stopErr.depsErrs) > 0 {
			s.failWithLock(stopErr)
		}
	})

	return s.err
}

func (s *Service) registerReadyCheck() error {
	if s.healthConfig.Name == "" {
		// if no name is set, use the service name
		s.healthConfig.Name = s.name
	}

	if s.healthConfig.Check != nil {
		s.healthConfig.Check = s.wrapCheck(s.healthConfig.Check)
	} else if checkable, ok := s.value.(Checkable); ok {
		s.healthConfig.Check = s.wrapCheck(checkable.Ready)
	} else {
		return nil
	}

	return s.app.readyHealth.Register(*s.healthConfig)
}

func (s *Service) wrapCheck(check health.CheckFunc) health.CheckFunc {
	return func(ctx context.Context) error {
		// we lock to make sure that the service is not
		// stopped while we are checking if it is ready
		s.mux.RLock()
		defer s.mux.RUnlock()

		switch s.Status() {
		case Constructing, Constructed:
			return fmt.Errorf("service not started")
		case Starting:
			return fmt.Errorf("service starting")
		case Running:
			return check(ctx)
		case Stopping:
			return fmt.Errorf("service stopping")
		case Stopped:
			return fmt.Errorf("service stopped")
		case Error:
			return fmt.Errorf("service in error state: %v", s.err)
		}
		return nil
	}
}

func (s *Service) registerMetric() {
	if collector, ok := s.value.(prometheus.Collector); ok {
		prometheus.WrapRegistererWithPrefix(s.metricsPrefix, s.app.prometheus).MustRegister(collector)
	}
}

type ServiceError struct {
	service *Service

	mu        sync.RWMutex
	directErr error
	depsErrs  []*ServiceError
}

func (e *ServiceError) Service() *Service {
	return e.service
}

func (e *ServiceError) Error() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var s string

	if e.directErr != nil {
		s = fmt.Sprintf("service %q: %v", e.service.name, e.directErr)
	} else {
		s = fmt.Sprintf("service %q", e.service.name)
	}

	if len(e.depsErrs) > 0 {
		for _, dep := range e.depsErrs {
			s += "\n"
			err := dep.Error()
			lines := strings.Split(err, "\n")
			indentedLines := make([]string, len(lines))
			for i, line := range lines {
				indentedLines[i] = fmt.Sprintf(">%s", line)
			}
			s += strings.Join(indentedLines, "\n")
		}
	}

	return s
}

func (e *ServiceError) addDepsErr(err *ServiceError) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.depsErrs = append(e.depsErrs, err)
}
