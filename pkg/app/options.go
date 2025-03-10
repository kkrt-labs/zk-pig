package app

import (
	"fmt"

	"github.com/hellofresh/health-go/v5"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

type Option func(*App) error

// WithAppName sets the name of the application.
func WithName(name string) Option {
	return func(a *App) error {
		a.name = name
		return nil
	}
}

// WithVersion sets the version of the application.
func WithVersion(version string) Option {
	return func(a *App) error {
		a.version = version
		return nil
	}
}

// WithLogger sets the logger of the application.
func WithLogger(logger *zap.Logger) Option {
	return func(a *App) error {
		a.logger = logger
		return nil
	}
}

type ServiceOption func(*Service) error

func WithMetricsPrefix(prefix string) ServiceOption {
	return func(s *Service) error {
		if !model.IsValidMetricName(model.LabelValue(prefix)) {
			return fmt.Errorf("invalid metrics prefix: %s (must contain only alphanumeric characters, underscores and colons)", prefix)
		}
		s.metricsPrefix = prefix
		return nil
	}
}

func WithHealthConfig(cfg *health.Config) ServiceOption {
	return func(s *Service) error {
		s.healthConfig = cfg
		return nil
	}
}
