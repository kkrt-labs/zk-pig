package log

import (
	"context"

	"github.com/kkrt-labs/kakarot-controller/pkg/tag"
	"go.uber.org/zap"
)

type loggerKey struct{}

// WithSugaredLogger returns a new context with the given sugared logger attached to it
func WithSugaredLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// LoggerWithFieldsFromContext returns a logger from the given context with the default namespace tags attached to it
func LoggerWithFieldsFromContext(ctx context.Context) *zap.Logger {
	return LoggerWithFieldsFromNamespaceContext(ctx, tag.DefaultNamespace)
}

// LoggerWithFieldsFromNamespaceContext returns a logger from the given context.
// It loads the tags from the provided tags namespace and adds them to the logger.
func LoggerWithFieldsFromNamespaceContext(ctx context.Context, namespaces ...string) *zap.Logger {
	if len(namespaces) == 0 {
		namespaces = []string{tag.DefaultNamespace}
	}

	logger := loggerFromContext(ctx)
	for _, namespace := range namespaces {
		tags := tag.FromNamespaceContext(ctx, namespace)
		fields := make([]zap.Field, 0, len(tags))
		for _, tag := range tags {
			fields = append(fields, zap.Any(string(tag.Key), tag.Value.Interface))
		}
		logger = logger.With(fields...)
	}
	return logger
}

// LoggerFromContext returns a logger from the given context with the default namespace tags attached to it
func LoggerFromContext(ctx context.Context) *zap.Logger {
	return LoggerWithFieldsFromNamespaceContext(ctx, tag.DefaultNamespace)
}

// loggerFromContext returns the logger attached to given context
func loggerFromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return logger
	}
	return zap.L()
}

// SugaredLoggerFromContext returns a sugared logger from the given context with the default namespace tags attached to it
func SugaredLoggerFromContext(ctx context.Context) *zap.SugaredLogger {
	return LoggerFromContext(ctx).Sugar()
}

// SugaredLoggerWithFieldsFromNamespaceContext returns a sugared logger from the given context.
// It loads the tags from the provided tags namespace and adds them to the logger.
func SugaredLoggerWithFieldsFromNamespaceContext(ctx context.Context, namespaces ...string) *zap.SugaredLogger {
	return LoggerWithFieldsFromNamespaceContext(ctx, namespaces...).Sugar()
}
