package cmd

import "github.com/spf13/cobra"

func NewRunCommand(rootCtx *RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the prover input generator daemon",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_ = rootCtx.App.Daemon()

			return rootCtx.App.Run(cmd.Context())
		},
	}

	return cmd
}
