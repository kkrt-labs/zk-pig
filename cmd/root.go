package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

// NewKKRTCtlCommand creates and returns the root command
func NewKKRTCtlCommand() *cobra.Command {
	var (
		logLevel  string
		logFormat string
	)

	rootCmd := &cobra.Command{
		Use:   "kkrtctl",
		Short: "kkrtctl is a CLI tool for managing prover inputs and more.",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if err := setupLogger(logLevel, logFormat); err != nil {
				return fmt.Errorf("failed to setup logger: %w", err)
			}
			return nil
		},
	}

	// Add persistent flags for logging
	pf := rootCmd.PersistentFlags()
	AddLogLevelFlag(&logLevel, pf)
	AddLogFormatFlag(&logFormat, pf)

	// Add subcommands
	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(NewProverInputsCommand())

	return rootCmd
}

func AddLogLevelFlag(logLevel *string, f *pflag.FlagSet) string {
	flagName := "log-level"
	f.StringVar(logLevel, flagName, "info", "Log level (debug|info|warn|error)")
	return flagName
}

func AddLogFormatFlag(logFormat *string, f *pflag.FlagSet) string {
	flagName := "log-format"
	f.StringVar(logFormat, flagName, "text", "Log format (json|text)")
	return flagName
}

func setupLogger(logLevel, logFormat string) error {
	cfg := zap.NewProductionConfig()

	// Log Level
	switch strings.ToLower(logLevel) {
	case "debug":
		cfg.Level.SetLevel(zap.DebugLevel)
	case "info":
		cfg.Level.SetLevel(zap.InfoLevel)
	case "warn":
		cfg.Level.SetLevel(zap.WarnLevel)
	case "error":
		cfg.Level.SetLevel(zap.ErrorLevel)
	case "":
		// do nothing, keep default from Production
	default:
		return fmt.Errorf("invalid log-level %q, must be one of: debug, info, warn, error", logLevel)
	}

	// Log Format
	if strings.EqualFold(logFormat, "text") {
		cfg.Encoding = "console"
	} else {
		cfg.Encoding = "json"
	}

	logger, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}
	zap.ReplaceGlobals(logger)
	return nil
}
