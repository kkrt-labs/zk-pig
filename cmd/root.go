package cmd

import (
	"github.com/spf13/cobra"
)

// NewKKRTCtlCommand creates and returns the root command
func NewKKRTCtlCommand() *cobra.Command {
	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "kkrtctl",
		Short: "kkrtctl is a CLI tool for managing prover inputs and more.",
	}

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(NewProverInputsCommand())

	return rootCmd
}
