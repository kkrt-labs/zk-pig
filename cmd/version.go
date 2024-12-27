package cmd

import (
	"fmt"

	"github.com/kkrt-labs/kakarot-controller/src"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("kkrtctl version %s\n", src.Version)
	},
}
