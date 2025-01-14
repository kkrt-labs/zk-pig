package log

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var levelsStr = []string{
	"debug",
	"info",
	"warn",
	"error",
}

var levels = map[string]Level{
	levelsStr[DebugLevel]: DebugLevel,
	levelsStr[InfoLevel]:  InfoLevel,
	levelsStr[WarnLevel]:  WarnLevel,
	levelsStr[ErrorLevel]: ErrorLevel,
}

func ParseLevel(level string) (Level, error) {
	if l, ok := levels[strings.ToLower(level)]; ok {
		return l, nil
	}
	return 0, fmt.Errorf("invalid log-level %q (must be one of %q)", level, levelsStr)
}

type Format int

const (
	TextFormat Format = iota
	JSONFormat
)

var formatsStr = []string{
	"text",
	"json",
}

var formats = map[string]Format{
	formatsStr[TextFormat]: TextFormat,
	formatsStr[JSONFormat]: JSONFormat,
}

func ParseFormat(format string) (Format, error) {
	if f, ok := formats[strings.ToLower(format)]; ok {
		return f, nil
	}
	return 0, fmt.Errorf("invalid log-format %q (must be one of %q)", format, formatsStr)
}

// NewLogger creates a new logger with the given log level and format.
func NewLogger(level Level, format Format) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	// Log Level
	switch level {
	case DebugLevel:
		cfg.Level.SetLevel(zap.DebugLevel)
	case InfoLevel:
		cfg.Level.SetLevel(zap.InfoLevel)
	case WarnLevel:
		cfg.Level.SetLevel(zap.WarnLevel)
	case ErrorLevel:
		cfg.Level.SetLevel(zap.ErrorLevel)
	default:
		return nil, fmt.Errorf("invalid log-level %q (must be one of %q)", level, levelsStr)
	}

	// Log Format
	switch format {
	case TextFormat:
		cfg.Encoding = "console"
	case JSONFormat:
		cfg.Encoding = "json"
	default:
		return nil, fmt.Errorf("invalid log-format %q (must be one of %q)", format, formatsStr)
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}
