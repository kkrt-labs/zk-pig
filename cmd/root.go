package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is our root command
// subcommands get attached here (version, prover-inputs, etc.)
var rootCmd = &cobra.Command{
	Use:   "kkrtctl",
	Short: "kkrtctl is a CLI tool for managing prover inputs and more.",
}

// Execute is called by main.go to run the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(proverInputsCmd)
}
