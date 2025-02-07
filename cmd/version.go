package cmd

import (
	"fmt"

	"github.com/kkrt-labs/zk-pig/src"
	"github.com/spf13/cobra"
)

func VersionCommand(_ *RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("%s\n", src.Version)
		},
	}
	return cmd
}
