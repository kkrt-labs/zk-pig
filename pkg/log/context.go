package log

import (
	"context"

	"github.com/kkrt-labs/kakarot-controller/pkg/tag"
	"go.uber.org/zap"
)

type loggerKey struct{}

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
		fields := TagsToFields(tags)
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

// WithLogger returns a new context with the given logger attached to it
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}
