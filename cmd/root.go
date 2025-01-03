package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
