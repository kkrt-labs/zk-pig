package src

import (
	"context"

	"github.com/kkrt-labs/go-utils/app"
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

func provide[T any](a *App, name string, constructor func() (T, error), opts ...app.ServiceOption) T {
	return app.Provide(a.app, name, constructor, opts...)
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
