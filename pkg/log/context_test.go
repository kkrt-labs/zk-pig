package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLoggerContext(t *testing.T) {
	// Create a test logger
	logger, _ := zap.NewDevelopment()
	
	// Create a context with the logger
	ctx := context.Background()
	ctxWithLogger := WithLogger(ctx, logger)
	
	// Test retrieving logger from context
	retrievedLogger := LoggerFromContext(ctxWithLogger)
	assert.Equal(t, logger, retrievedLogger, "retrieved logger should match original logger")
	
	// Test with empty context
	emptyCtxLogger := LoggerFromContext(context.Background())
	assert.Equal(t, zap.L(), emptyCtxLogger, "empty context should return global logger")
} 
