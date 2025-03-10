package app

import (
	"go.uber.org/zap"
)

type Option func(*App) error

// WithName sets the name of the application.
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
