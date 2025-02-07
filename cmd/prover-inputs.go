package cmd

import (
	"fmt"
	"math/big"

	"github.com/kkrt-labs/go-utils/ethereum/rpc/jsonrpc"
	"github.com/kkrt-labs/zk-pig/src"
	"github.com/kkrt-labs/zk-pig/src/config"
	"github.com/spf13/cobra"
)

type ProverInputContext struct {
	RootContext
	svc         *src.Service
	blockNumber *big.Int
}

// NewGenerateCommand creates and returns the generate command
func NewGenerateCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputContext{RootContext: *rootCtx}
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate prover input for a specific block",
		Long:    "Generate prover inputs by running preflight, prepare and execute in a single run. It runs online and requires --chain-rpc-url to be set to a remote JSON-RPC Ethereum Execution Layer node",
		PreRunE: preRun(ctx, &blockNumber),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Generate(cmd.Context(), ctx.blockNumber)
		},
		PostRunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Stop(cmd.Context())
		},
	}

	config.AddProverInputFlags(ctx.Viper, cmd.PersistentFlags())
	cmd.Flags().StringVarP(&blockNumber, "block-number", "b", "latest", "Block number")

	return cmd
}

func NewPreflightCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputContext{RootContext: *rootCtx}
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:     "preflight",
		Short:   "Collect necessary data to generate prover inputs from a remote JSON-RPC Ethereum Execution Layer node",
		Long:    "Collect necessary data to generate prover inputs from a remote JSON-RPC Ethereum Execution Layer node. It runs online and requires --chain-rpc-url to be set to a remote JSON-RPC Ethereum Execution Layer node",
		PreRunE: preRun(ctx, &blockNumber),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Preflight(cmd.Context(), ctx.blockNumber)
		},
		PostRunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Stop(cmd.Context())
		},
	}

	config.AddProverInputFlags(ctx.Viper, cmd.PersistentFlags())
	cmd.Flags().StringVarP(&blockNumber, "block-number", "b", "latest", "Block number")

	return cmd
}

func NewPrepareCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputContext{RootContext: *rootCtx}
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:     "prepare",
		Short:   "Prepare prover inputs by basing on data previously collected during preflight.",
		Long:    "Prepare prover inputs by basing on data previously collected during preflight. It can be ran off-line in which case it needs --chain-id to be provided",
		PreRunE: preRun(ctx, &blockNumber),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Prepare(cmd.Context(), ctx.blockNumber)
		},
		PostRunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Stop(cmd.Context())
		},
	}

	config.AddProverInputFlags(ctx.Viper, cmd.PersistentFlags())
	cmd.Flags().StringVarP(&blockNumber, "block-number", "b", "latest", "Block number")

	return cmd
}

func NewExecuteCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputContext{RootContext: *rootCtx}
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:     "execute",
		Short:   "Execute block by basing on prover inputs previously generated during prepare.",
		Long:    "Execute block by basing on prover inputs previously generated during prepare. It can be ran off-line in which case it needs --chain-id to be provided.",
		PreRunE: preRun(ctx, &blockNumber),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Execute(cmd.Context(), ctx.blockNumber)
		},
		PostRunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Stop(cmd.Context())
		},
	}

	config.AddProverInputFlags(ctx.Viper, cmd.PersistentFlags())
	cmd.Flags().StringVarP(&blockNumber, "block-number", "b", "latest", "Block number")

	return cmd
}

func preRun(ctx *ProverInputContext, blockNumber *string) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		var err error
		ctx.svc, err = src.FromGlobalConfig(ctx.Config)
		if err != nil {
			return fmt.Errorf("failed to create prover inputs service: %v", err)
		}

		err = ctx.svc.Start(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to start prover inputs service: %v", err)
		}

		ctx.blockNumber, err = jsonrpc.FromBlockNumArg(*blockNumber)
		if err != nil {
			return fmt.Errorf("invalid block number: %v", err)
		}

		if err := validateS3Config(ctx); err != nil {
			return err
		}

		return nil
	}
}

// Helper function to validate S3 configuration
func validateS3Config(ctx *ProverInputContext) error {
	if ctx.Config.ProverInputtore.S3.AWSProvider.Bucket == "" || ctx.Config.ProverInputtore.S3.AWSProvider.KeyPrefix == "" || ctx.Config.ProverInputtore.S3.AWSProvider.Credentials.AccessKey == "" || ctx.Config.ProverInputtore.S3.AWSProvider.Credentials.SecretKey == "" || ctx.Config.ProverInputtore.S3.AWSProvider.Region == "" {
		missingFields := []string{}
		if ctx.Config.ProverInputtore.S3.AWSProvider.Bucket == "" {
			missingFields = append(missingFields, "s3-bucket")
		}
		if ctx.Config.ProverInputtore.S3.AWSProvider.KeyPrefix == "" {
			missingFields = append(missingFields, "key-prefix")
		}
		if ctx.Config.ProverInputtore.S3.AWSProvider.Credentials.AccessKey == "" {
			missingFields = append(missingFields, "access-key")
		}
		if ctx.Config.ProverInputtore.S3.AWSProvider.Credentials.SecretKey == "" {
			missingFields = append(missingFields, "secret-key")
		}
		if ctx.Config.ProverInputtore.S3.AWSProvider.Region == "" {
			missingFields = append(missingFields, "region")
		}
		return fmt.Errorf("%s must be specified when using s3 storage", missingFields)
	}
	return nil
}
