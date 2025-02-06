package cmd

import (
	"fmt"
	"math/big"

	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/jsonrpc"
	"github.com/kkrt-labs/kakarot-controller/src/blocks"
	"github.com/kkrt-labs/kakarot-controller/src/config"
	"github.com/spf13/cobra"
)

type ProverInputsContext struct {
	RootContext
	svc         *blocks.Service
	blockNumber *big.Int
}

// 1. Main command
func NewProverInputsCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputsContext{RootContext: *rootCtx}
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:   "prover-inputs",
		Short: "Commands for generating and validating prover inputs",
		RunE:  runHelp,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			var err error
			ctx.svc, err = blocks.FromGlobalConfig(ctx.Config)
			if err != nil {
				return fmt.Errorf("failed to create prover inputs service: %v", err)
			}

			err = ctx.svc.Start(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to start prover inputs service: %v", err)
			}

			ctx.blockNumber, err = jsonrpc.FromBlockNumArg(blockNumber)
			if err != nil {
				return fmt.Errorf("invalid block number: %v", err)
			}

			if ctx.Config.Store.Location == "s3" {
				if ctx.Config.AWS.S3.Bucket == "" {
					return fmt.Errorf("s3-bucket must be specified when using s3 storage")
				}
				if ctx.Config.AWS.S3.KeyPrefix == "" {
					return fmt.Errorf("key-prefix must be specified when using s3 storage")
				}
				if ctx.Config.AWS.S3.AccessKey == "" {
					return fmt.Errorf("access-key must be specified when using s3 storage")
				}
				if ctx.Config.AWS.S3.SecretKey == "" {
					return fmt.Errorf("secret-key must be specified when using s3 storage")
				}
				if ctx.Config.AWS.S3.Region == "" {
					return fmt.Errorf("region must be specified when using s3 storage")
				}
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Stop(cmd.Context())
		},
	}

	config.AddProverInputsFlags(ctx.Viper, cmd.PersistentFlags())

	cmd.PersistentFlags().StringVarP(&blockNumber, "block-number", "b", "latest", "Block number")

	cmd.AddCommand(
		NewGenerateCommand(ctx),
		NewPreflightCommand(ctx),
		NewPrepareCommand(ctx),
		NewExecuteCommand(ctx),
	)

	return cmd
}

func runHelp(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

// 2. Subcommands
func NewGenerateCommand(ctx *ProverInputsContext) *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate prover inputs",
		Long:  "Generate prover inputs by running preflight, prepare and execute in a single run",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Generate(cmd.Context(), ctx.blockNumber)
		},
	}
}

func NewPreflightCommand(ctx *ProverInputsContext) *cobra.Command {
	return &cobra.Command{
		Use:   "preflight",
		Short: "Collect necessary data for proving a block from a remote RPC node",
		Long:  "Collect necessary data for proving a block from a remote RPC node. It processes the EVM block on a state and chain which database have been replaced with a connector to a remote JSON-RPC node",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Preflight(cmd.Context(), ctx.blockNumber)
		},
	}
}

func NewPrepareCommand(ctx *ProverInputsContext) *cobra.Command {
	return &cobra.Command{
		Use:   "prepare",
		Short: "Prepare prover inputs, basing on data collected during preflight",
		Long:  "Prepare prover inputs, basing on data collected during preflight. It processes and validates an EVM block over in memory state and chain prefilled with data collected during preflight.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Prepare(cmd.Context(), ctx.blockNumber)
		},
	}
}

func NewExecuteCommand(ctx *ProverInputsContext) *cobra.Command {
	return &cobra.Command{
		Use:   "execute",
		Short: "Run an EVM execution, basing on prover inputs generated during prepare",
		Long:  "Run an EVM execution, basing on prover inputs generated during prepare. It processes and validates an EVM block over in memory state and chain prefilled with prover inputs.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return ctx.svc.Execute(cmd.Context(), ctx.blockNumber)
		},
	}
}
