package src

import (
	"context"
	"fmt"

	"github.com/kkrt-labs/go-utils/app"
	"github.com/kkrt-labs/go-utils/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Service is a service that enables the generation of prover inpunts for EVM compatible blocks.
type App struct {
	app *app.App
	cfg *Config
}

func NewApp(cfg *Config) (*App, error) {
	a, err := app.NewApp(
		cfg.App,
		app.WithName("zkpig"),
		app.WithVersion(Version),
	)
	if err != nil {
		return nil, err
	}

	return &App{
		app: a,
		cfg: cfg,
	}, nil
}

func (a *App) Config() *Config {
	return a.cfg
}

// Load loads the app from default viper config
func Load() (*App, error) {
	return LoadFromViper(config.NewViper())
}

// LoadFromViper loads the app from viper
func LoadFromViper(v *viper.Viper) (*App, error) {
	_ = AddFlags(v, pflag.NewFlagSet("zkpig", pflag.ContinueOnError))

	cfg := new(Config)
	err := cfg.Load(v)
	if err != nil {
		return nil, fmt.Errorf("failed to load zkpig configuration: %w", err)
	}

	app, err := NewApp(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create zkpig app: %w", err)
	}

	return app, nil
}

func (a *App) Context(ctx context.Context) context.Context {
	return a.app.Context(ctx)
}

func (a *App) Start(ctx context.Context) error {
	return a.app.Start(ctx)
}

func (a *App) Stop(ctx context.Context) error {
	return a.app.Stop(ctx)
}

func (a *App) Run(ctx context.Context) error {
	return a.app.Run(ctx)
}

func (a *App) Error() error {
	return a.app.Error()
}

func provide[T any](a *App, name string, constructor func() (T, error), opts ...app.ServiceOption) T {
	return app.Provide(a.app, name, constructor, opts...)
}
