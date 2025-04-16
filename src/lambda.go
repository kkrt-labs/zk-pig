package src

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

// StartLambdaWithApp is an utility function to facilitate the creation of a lambda function with an app.
func StartLambdaWithApp(ctx context.Context, createHandler func(ctx context.Context, app *App) any, opts ...lambda.Option) error {
	app, err := Load()
	if err != nil {
		return err
	}

	handler := createHandler(ctx, app)

	err = app.Start(ctx)
	if err != nil {
		return err
	}

	opts = append(
		opts,
		lambda.WithContext(app.Context(ctx)),
		lambda.WithEnableSIGTERM(func() {
			err = app.Stop(ctx)
		}),
	)

	lambda.StartWithOptions(handler, opts...)

	return err
}
