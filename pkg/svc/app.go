package svc

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"go.uber.org/zap"
)

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

	status atomic.Uint32

	err *ServiceError

	startOnce sync.Once

	stopOnce sync.Once
	stopChan chan struct{}
}

func newService(name string, constructor func() (any, error)) *Service {
	return &Service{
		name:        name,
		constructor: constructor,
		deps:        make(map[string]*Service),
		depsOf:      make(map[string]*Service),
		stopChan:    make(chan struct{}),
	}
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

func (s *Service) construct() {
	s.setStatus(Constructing)
	svc, constructorErr := s.constructor()
	if constructorErr != nil {
		_ = s.fail(constructorErr)
		return
	}

	s.value = svc

	s.setStatus(Constructed)
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

func (s *Service) start(ctx context.Context) *ServiceError {
	s.startOnce.Do(func() {
		if s.err != nil {
			return
		}

		s.setStatus(Starting)

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
			_ = s.fail(startErr)
			return
		}

		// If all dependencies started successfully then start the service
		if s.err == nil {
			if start, ok := s.value.(Runnable); ok {
				s.app.logger.Info("Starting service", zap.String("service", s.name))
				err := start.Start(ctx)
				if err != nil {
					_ = s.fail(err)
					s.app.logger.Error("Failed to start service", zap.String("service", s.name), zap.Error(err))
					return
				}
				s.app.logger.Info("Service started", zap.String("service", s.name))
			}
		}

		s.setStatus(Running)
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

		s.setStatus(Stopping)
		defer func() {
			close(s.stopChan)
		}()

		if stop, ok := s.value.(Runnable); ok {
			s.app.logger.Info("Stopping service", zap.String("service", s.name))
			err := stop.Stop(ctx)
			if err != nil {
				_ = s.fail(err)
				s.app.logger.Error("Failed to stop service", zap.String("service", s.name), zap.Error(err))
				return
			}
			s.app.logger.Info("Service stopped", zap.String("service", s.name))
		}
		if s.err == nil {
			s.setStatus(Stopped)
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
			_ = s.fail(stopErr)
		}
	})

	return s.err
}

type App struct {
	services map[string]*Service

	top     *Service
	current *Service

	done chan os.Signal

	logger *zap.Logger
}

func NewApp(logger *zap.Logger) *App {
	return &App{
		services: make(map[string]*Service),
		done:     make(chan os.Signal),
		logger:   logger,
	}
}

func (app *App) Provide(name string, constructor func() (any, error)) any {
	if name == "" {
		name = reflect.TypeOf(constructor).Out(0).String()
	}

	if svc, ok := app.services[name]; ok {
		app.current.addDep(svc) // current can not be nil here
		return svc.value
	}

	svc := app.addService(name, constructor)

	return svc.value
}

func (app *App) addService(name string, constructor func() (any, error)) *Service {
	previous := app.current
	svc := newService(name, constructor)
	svc.app = app
	app.services[name] = svc
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

func Provide[T any](app *App, name string, constructor func() (T, error)) T {
	val := app.Provide(name, func() (any, error) {
		return constructor()
	})

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
