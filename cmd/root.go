package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/kkrt-labs/go-utils/config"
	"github.com/kkrt-labs/zk-pig/src"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.EnableTraverseRunHooks = true
}

type RootContext struct {
	Viper  *viper.Viper
	Config *src.Config
	App    *src.App
}

// NewZkPigCommand creates and returns the root command
func NewZkPigCommand() *cobra.Command {
	ctx := &RootContext{
		Viper: config.NewViper(),
	}

	rootCmd := &cobra.Command{
		Use:   "zkpig",
		Short: "zkpig is an application for generating ZK-EVM prover inputs.",
		PersistentPreRunE: func(rootCmd *cobra.Command, _ []string) error {
			cfg := new(src.Config)
			err := cfg.Load(ctx.Viper)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			ctx.Config = cfg

			ctx.App, err = src.NewApp(ctx.Config)
			if err != nil {
				return fmt.Errorf("failed to create app: %w", err)
			}

			rootCmd.SetContext(ctx.App.Context(rootCmd.Context()))

			return nil
		},
	}

	// Add persistent flags for logging
	_ = src.AddFlags(ctx.Viper, rootCmd.PersistentFlags())

	// Add subcommands
	rootCmd.AddCommand(VersionCommand(ctx))
	rootCmd.AddCommand(NewGenerateCommand(ctx))
	rootCmd.AddCommand(NewPreflightCommand(ctx))
	rootCmd.AddCommand(NewPrepareCommand(ctx))
	rootCmd.AddCommand(NewExecuteCommand(ctx))
	rootCmd.AddCommand(NewRunCommand(ctx))
	rootCmd.AddCommand(NewConfigCommand(ctx))

	return rootCmd
}

func NewConfigCommand(rootCtx *RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Returns current configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(rootCtx.Config)
		},
	}

	return cmd
}
