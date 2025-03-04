package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/kkrt-labs/go-utils/log"
	"github.com/kkrt-labs/zk-pig/src"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
		Viper: viper.New(),
	}

	rootCmd := &cobra.Command{
		Use:   "zkpig",
		Short: "zkpig is a CLI tool for generating and validating ZK-EVM prover inputs.",
		PersistentPreRunE: func(rootCmd *cobra.Command, _ []string) error {
			ctx.Config = new(src.Config)
			if err := ctx.Config.Load(ctx.Viper); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			logCfg, err := log.ParseConfig(&ctx.Config.Log)
			if err != nil {
				return err
			}

			logCfg.OutputPaths = []string{"stderr"}
			logCfg.ErrorOutputPaths = []string{"stderr"}

			logger, err := logCfg.Build()
			if err != nil {
				return fmt.Errorf("failed to create logger: %w", err)
			}

			logger.Info("Starting zkpig", zap.Any("config", ctx.Config))

			err = validateS3Config(ctx.Config)
			if err != nil {
				return err
			}

			app, err := src.NewApp(ctx.Config, logger)
			if err != nil {
				return fmt.Errorf("failed to create app: %w", err)
			}
			ctx.App = app

			rootCmd.SetContext(log.WithLogger(rootCmd.Context(), logger))

			return nil
		},
	}

	// Add persistent flags for logging
	log.AddFlags(ctx.Viper, rootCmd.PersistentFlags())
	src.AddConfigFileFlag(ctx.Viper, rootCmd.PersistentFlags())

	// Add flags for chain, aws, and store
	src.AddChainFlags(ctx.Viper, rootCmd.PersistentFlags())
	src.AddAWSFlags(ctx.Viper, rootCmd.PersistentFlags())
	src.AddStoreFlags(ctx.Viper, rootCmd.PersistentFlags())

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

// Helper function to validate S3 configuration
func validateS3Config(cfg *src.Config) error {
	if cfg == nil {
		return fmt.Errorf("config is not set")
	}

	// Check if any S3 field is set
	if cfg.ProverInputStore.S3.Bucket != "" ||
		cfg.ProverInputStore.S3.BucketKeyPrefix != "" ||
		cfg.ProverInputStore.S3.AWSProvider.Credentials.AccessKey != "" ||
		cfg.ProverInputStore.S3.AWSProvider.Credentials.SecretKey != "" ||
		cfg.ProverInputStore.S3.AWSProvider.Region != "" {

		// If any S3 field is set, ensure all required fields are set
		missingFields := []string{}
		if cfg.ProverInputStore.S3.Bucket == "" {
			missingFields = append(missingFields, "s3-bucket")
		}
		if cfg.ProverInputStore.S3.AWSProvider.Credentials.AccessKey == "" {
			missingFields = append(missingFields, "access-key")
		}
		if cfg.ProverInputStore.S3.AWSProvider.Credentials.SecretKey == "" {
			missingFields = append(missingFields, "secret-key")
		}
		if cfg.ProverInputStore.S3.AWSProvider.Region == "" {
			missingFields = append(missingFields, "region")
		}

		// If any required field is missing, return an error
		if len(missingFields) > 0 {
			return fmt.Errorf("%s must be specified when using s3 storage", missingFields)
		}
	}

	return nil
}
