package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestApp(t *testing.T) *App {
	cfg := new(Config)
	cfg.Main.Entrypoint.Network = "tcp"
	cfg.Main.Entrypoint.KeepAlive = "1s"
	cfg.Main.ReadTimeout = "1s"
	cfg.Main.ReadHeaderTimeout = "1s"
	cfg.Main.WriteTimeout = "1s"
	cfg.Main.IdleTimeout = "1s"
	cfg.Healthz.Entrypoint.Network = "tcp"
	cfg.Healthz.Entrypoint.KeepAlive = "1s"
	cfg.Healthz.ReadTimeout = "1s"
	cfg.Healthz.ReadHeaderTimeout = "1s"
	cfg.Healthz.WriteTimeout = "1s"
	cfg.Healthz.IdleTimeout = "1s"

	app, err := NewApp(
		cfg,
		WithLogger(zap.NewNop()),
		WithName("test"),
		WithVersion("1.0.0"),
	)
	require.NoError(t, err)
	return app
}

func TestAppProvide(t *testing.T) {
	var testCase = []struct {
		desc        string
		constructor func() (any, error)
		expected    any
		expectErr   bool
	}{
		{
			desc: "string",
			constructor: func() (any, error) {
				return "test", nil
			},
			expected: "test",
		},
		{
			desc: "int",
			constructor: func() (any, error) {
				return 1, nil
			},
			expected: 1,
		},
		{
			desc: "nil",
			constructor: func() (any, error) {
				return nil, nil
			},
			expected: nil,
		},
		{
			desc: "error",
			constructor: func() (any, error) {
				return nil, errors.New("error")
			},
			expectErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.desc, func(t *testing.T) {
			app := newTestApp(t)
			res := app.Provide("test", tc.constructor)
			err := app.Error()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, res)
			}
		})
	}
}

func TestProvide(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (string, error) {
			return "test", nil
		})
		assert.Equal(t, res, "test")
	})

	t.Run("int", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (int, error) {
			return 1, nil
		})
		assert.Equal(t, res, 1)
	})

	t.Run("*string", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (*string, error) {
			return nil, nil
		})
		assert.Nil(t, res)
	})

	t.Run("*string#nil", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (*string, error) {
			return nil, nil
		})
		assert.Nil(t, res)
	})

	t.Run("error", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (error, error) {
			return errors.New("error"), nil
		})
		assert.Equal(t, errors.New("error"), res)
	})

	t.Run("interface", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (interface{}, error) {
			return "test", nil
		})
		assert.Equal(t, res, "test")
	})

	t.Run("interface#nil", func(t *testing.T) {
		app := newTestApp(t)
		res := Provide(app, "test", func() (interface{}, error) {
			return nil, nil
		})
		assert.Nil(t, res)
	})
}

type testService struct {
	start chan error
	stop  chan error
}

func (s *testService) Start(_ context.Context) error {
	return <-s.start
}

func (s *testService) Stop(_ context.Context) error {
	return <-s.stop
}

func TestAppNoDeps(t *testing.T) {
	start, stop := make(chan error), make(chan error)
	defer close(start)
	defer close(stop)

	testApp := func() *App {
		app := newTestApp(t)
		_ = Provide(app, "test", func() (*testService, error) {
			return &testService{
				start: start,
				stop:  stop,
			}, nil
		})
		return app
	}

	recStart, recStop := make(chan error), make(chan error)
	defer close(recStart)
	defer close(recStop)

	t.Run("no errors", func(t *testing.T) {
		app := testApp()
		require.Equal(t, app.services["test"].Status(), Constructed)

		go func() {
			recStart <- app.Start(context.Background())
		}()
		time.Sleep(100 * time.Millisecond) // wait for the service to start
		assert.Equal(t, app.services["test"].Status(), Starting)

		// Trigger start
		start <- nil
		assert.Nil(t, <-recStart)
		assert.Equal(t, app.services["test"].Status(), Running)

		go func() {
			recStop <- app.Stop(context.Background())
		}()
		time.Sleep(100 * time.Millisecond) // wait for the service to start
		assert.Equal(t, app.services["test"].Status(), Stopping)

		// Trigger stop
		stop <- nil
		assert.Nil(t, <-recStop)
		assert.Equal(t, app.services["test"].Status(), Stopped)
	})

	t.Run("error on start", func(t *testing.T) {
		app := testApp()
		go func() {
			recStart <- app.Start(context.Background())
		}()

		start <- errors.New("error on start")
		startErr := <-recStart
		require.IsType(t, startErr, &ServiceError{})
		assert.Equal(t, startErr.(*ServiceError).directErr, errors.New("error on start"))
		assert.Equal(t, app.services["test"].Status(), Error)
	})

	t.Run("error on stop", func(t *testing.T) {
		app := testApp()
		go func() {
			recStart <- app.Start(context.Background())
		}()
		start <- nil
		<-recStart

		go func() {
			recStop <- app.Stop(context.Background())
		}()
		stop <- errors.New("error on stop")
		stopErr := <-recStop
		require.IsType(t, stopErr, &ServiceError{})
		assert.Equal(t, stopErr.(*ServiceError).directErr, errors.New("error on stop"))
		assert.Equal(t, app.services["test"].Status(), Error)
	})
}

func TestAppWithDeps(t *testing.T) {
	app := newTestApp(t)
	startMain, stopMain, startDep, stopDep := make(chan error), make(chan error), make(chan error), make(chan error)
	defer close(startMain)
	defer close(stopMain)
	defer close(startDep)
	defer close(stopDep)

	_ = Provide(app, "main", func() (*testService, error) {
		_ = Provide(app, "dep", func() (*testService, error) {
			return &testService{
				start: startDep,
				stop:  stopDep,
			}, nil
		})
		return &testService{
			start: startMain,
			stop:  stopMain,
		}, nil
	})

	// Test dependency tree
	assert.Equal(t, app.services["main"].deps["dep"], app.services["dep"])
	assert.Equal(t, app.services["dep"].depsOf["main"], app.services["main"])

	recStart, recStop := make(chan error), make(chan error)
	defer close(recStart)
	defer close(recStop)

	go func() {
		recStart <- app.Start(context.Background())
	}()
	startDep <- nil
	startMain <- nil
	assert.Nil(t, <-recStart)

	go func() {
		recStop <- app.Stop(context.Background())
	}()
	stopMain <- nil
	stopDep <- nil
	assert.Nil(t, <-recStop)
}

func TestServiceError(t *testing.T) {
	svc := newService("svc", nil)
	dep1 := newService("dep1", nil)
	dep2 := newService("dep2", nil)
	dep11 := newService("dep11", nil)
	dep21 := newService("dep21", nil)

	svcErr := svc.fail(errors.New("error on svc"))
	dep1Err := dep1.fail(nil)
	dep2Err := dep2.fail(errors.New("error on dep2"))
	dep1Err.addDepsErr(dep11.fail(errors.New("error on dep11")))
	dep2Err.addDepsErr(dep21.fail(nil))

	svcErr.addDepsErr(dep1Err)
	svcErr.addDepsErr(dep2Err)

	expected := `service "svc": error on svc
>service "dep1"
>>service "dep11": error on dep11
>service "dep2": error on dep2
>>service "dep21"`
	assert.Equal(t, svcErr.Error(), expected)
}

func TestAppServers(t *testing.T) {
	app := newTestApp(t)
	require.NoError(t, app.Error())

	app.Provide("top", func() (any, error) {
		app.EnableMain()
		app.EnableHealthz()
		return nil, nil
	})

	err := app.Start(context.Background())
	require.NoError(t, err)

	// Check main server is running
	require.NotNil(t, app.main)
	mainAddr := app.main.Addr()
	require.NotEmpty(t, mainAddr)

	conn, err := net.Dial("tcp", mainAddr)
	require.NoError(t, err)
	conn.Close()

	// Check main server is running
	require.NotNil(t, app.healthz)
	healthzAddr := app.healthz.Addr()
	require.NotEmpty(t, healthzAddr)

	conn, err = net.Dial("tcp", healthzAddr)
	require.NoError(t, err)
	conn.Close()

	// Check healthz server is running
	err = app.Stop(context.Background())
	require.NoError(t, err)
}

type checkableService struct {
	err error
}

func (s *checkableService) Ready(_ context.Context) error {
	return s.err
}

func TestHealthChecks(t *testing.T) {
	app := newTestApp(t)
	require.NoError(t, app.Error())

	checkable := new(checkableService)
	Provide(app, "checkable", func() (*checkableService, error) {
		app.EnableHealthz()
		return checkable, nil
	})

	err := app.Start(context.Background())
	require.NoError(t, err)

	require.NotNil(t, app.healthz)
	healthAddr := app.healthz.Addr()
	require.NotEmpty(t, healthAddr)

	// Test live check
	req, err := http.NewRequest("GET", "http://"+healthAddr+"/live", http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	// Test ready check
	req, err = http.NewRequest("GET", "http://"+healthAddr+"/ready", http.NoBody)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	// Test ready check with error
	checkable.err = errors.New("test error")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusServiceUnavailable)
}

type metricsService struct {
	count prometheus.Counter
}

func (s *metricsService) incr() {
	s.count.Inc()
}

func (s *metricsService) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.count.Desc()
}

func (s *metricsService) Collect(ch chan<- prometheus.Metric) {
	ch <- s.count
}

func TestMetrics(t *testing.T) {
	app := newTestApp(t)
	require.NoError(t, app.Error())

	metrics := &metricsService{
		count: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_metricA",
		}),
	}
	app.Provide("test-metrics-svc", func() (any, error) {
		app.EnableHealthz()
		app.Provide(
			"test-metrics-svc-with-prefix",
			func() (any, error) {
				return &metricsService{
					count: prometheus.NewCounter(prometheus.CounterOpts{
						Name: "test_metricB",
					}),
				}, nil
			},
			WithMetricsPrefix("test_prefix_"),
		)
		return metrics, nil
	})

	err := app.Start(context.Background())
	require.NoError(t, err)

	require.NotNil(t, app.healthz)
	healthAddr := app.healthz.Addr()
	require.NotEmpty(t, healthAddr)

	// Test metrics endpoint
	req, err := http.NewRequest("GET", "http://"+healthAddr+"/metrics", http.NoBody)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	// Test collectors are registered with correct labels
	families, err := app.prometheus.Gather()
	require.NoError(t, err)
	familyCount := len(families)
	assert.GreaterOrEqual(t, familyCount, 2)
	assert.Equal(t, "test_metricA", families[familyCount-2].GetName())
	assert.Equal(t, "test_prefix_test_metricB", families[familyCount-1].GetName())

	// Test metrics are updated
	assert.Equal(t, float64(0), families[familyCount-2].GetMetric()[0].GetCounter().GetValue())
	metrics.incr()
	metrics.incr()
	metrics.incr()

	families, err = app.prometheus.Gather()
	require.NoError(t, err)
	assert.Equal(t, float64(3), families[familyCount-2].GetMetric()[0].GetCounter().GetValue())

	err = app.Stop(context.Background())
	require.NoError(t, err)
}
