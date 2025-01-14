package cmd

import (
	"fmt"

	"github.com/kkrt-labs/kakarot-controller/pkg/log"
	"github.com/kkrt-labs/kakarot-controller/src/config"
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

// NewKKRTCtlCommand creates and returns the root command
func NewKKRTCtlCommand() *cobra.Command {
	ctx := &RootContext{
		Viper:  viper.New(),
		Config: new(config.Config),
	}

	rootCmd := &cobra.Command{
		Use:   "kkrtctl",
		Short: "kkrtctl is a CLI tool for managing prover inputs and more.",
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

	// Add subcommands
	rootCmd.AddCommand(VersionCommand(ctx))
	rootCmd.AddCommand(NewProverInputsCommand(ctx))

	return rootCmd
}
