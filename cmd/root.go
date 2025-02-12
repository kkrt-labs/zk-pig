package cmd

import (
	"fmt"

	"github.com/kkrt-labs/go-utils/log"
	"github.com/kkrt-labs/zk-pig/src/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.EnableTraverseRunHooks = true
}

type RootContext struct {
	Config *config.Config
	Viper  *viper.Viper
}

// NewZkPigCommand creates and returns the root command
func NewZkPigCommand() *cobra.Command {
	ctx := &RootContext{
		Viper:  viper.New(),
		Config: new(config.Config),
	}

	rootCmd := &cobra.Command{
		Use:   "zkpig",
		Short: "zkpig is a CLI tool for generating and validating ZK-EVM prover inputs.",
		PersistentPreRunE: func(rootCmd *cobra.Command, _ []string) error {
			if err := ctx.Config.Load(ctx.Viper); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			level, err := log.ParseLevel(ctx.Config.Log.Level)
			if err != nil {
				return err
			}

			format, err := log.ParseFormat(ctx.Config.Log.Format)
			if err != nil {
				return err
			}

			logger, err := log.NewLogger(level, format)
			if err != nil {
				return fmt.Errorf("failed to create logger: %w", err)
			}

			ctx := log.WithLogger(rootCmd.Context(), logger)
			rootCmd.SetContext(ctx)

			return nil
		},
	}

	// Add persistent flags for logging
	log.AddFlags(ctx.Viper, rootCmd.PersistentFlags())
	config.AddConfigFileFlag(ctx.Viper, rootCmd.PersistentFlags())

	// Add flags for chain, aws, and store
	config.AddChainFlags(ctx.Viper, rootCmd.PersistentFlags())
	config.AddAWSFlags(ctx.Viper, rootCmd.PersistentFlags())
	config.AddStoreFlags(ctx.Viper, rootCmd.PersistentFlags())

	// Add subcommands
	rootCmd.AddCommand(VersionCommand(ctx))
	rootCmd.AddCommand(NewGenerateCommand(ctx))
	rootCmd.AddCommand(NewPreflightCommand(ctx))
	rootCmd.AddCommand(NewPrepareCommand(ctx))
	rootCmd.AddCommand(NewExecuteCommand(ctx))
	rootCmd.AddCommand(NewConfigCommand(ctx))

	return rootCmd
}
