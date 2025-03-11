package app

import (
	"github.com/hellofresh/health-go/v5"
	"github.com/kkrt-labs/go-utils/tag"
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

type ServiceOption func(*service) error

func WithHealthConfig(cfg *health.Config) ServiceOption {
	return func(s *service) error {
		if cfg.Name != "" {
			s.healthConfig.Name = cfg.Name
		}

		if cfg.Check != nil {
			s.healthConfig.Check = s.wrapCheck(cfg.Check)
		}

		if cfg.Timeout != 0 {
			s.healthConfig.Timeout = cfg.Timeout
		}

		if cfg.SkipOnErr {
			s.healthConfig.SkipOnErr = true
		}

		return nil
	}
}

func WithTags(tags ...*tag.Tag) ServiceOption {
	return func(s *service) error {
		s.tags = s.tags.WithTags(tags...)
		return nil
	}
}

// WithComponentName sets the name of the component.
// Multiple services can have the same component name.
// By default, the component name is the name of the service identifier which is unique.
// This enables to override the default name (without unicity constraints)
func WithComponentName(name string) ServiceOption {
	return func(s *service) error {
		s.name = name
		return WithTags(tag.Key("component").String(name))(s)
	}
}
