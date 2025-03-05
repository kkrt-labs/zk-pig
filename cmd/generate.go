package cmd

import (
	"fmt"
	"math/big"

	"github.com/kkrt-labs/go-utils/ethereum/rpc/jsonrpc"
	"github.com/spf13/cobra"
)

type ProverInputContext struct {
	*RootContext
	blockNumber *big.Int
}

// NewGenerateCommand creates and returns the generate command
func NewGenerateCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputContext{RootContext: rootCtx}
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate prover input for a specific block",
		Long:    "Generate prover inputs by running preflight, prepare and execute in a single run. It runs online and requires --chain-rpc-url to be set to a remote JSON-RPC Ethereum Execution Layer node",
		PreRunE: preRun(ctx, &blockNumber),
		RunE: func(cmd *cobra.Command, _ []string) error {
			generator := ctx.App.Generator() // must be declared first so object is constructed on App before calling Start

			defer ctx.App.Stop(cmd.Context())
			err := ctx.App.Start(cmd.Context())
			if err != nil {
				return err
			}

			return generator.Generate(cmd.Context(), ctx.blockNumber)
		},
	}

	cmd.Flags().StringVarP(&blockNumber, "block-number", "b", "latest", "Block number")

	return cmd
}

func NewPreflightCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputContext{RootContext: rootCtx}
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:     "preflight",
		Short:   "Collect necessary data to generate prover inputs from a remote JSON-RPC Ethereum Execution Layer node",
		Long:    "Collect necessary data to generate prover inputs from a remote JSON-RPC Ethereum Execution Layer node. It runs online and requires --chain-rpc-url to be set to a remote JSON-RPC Ethereum Execution Layer node",
		PreRunE: preRun(ctx, &blockNumber),
		RunE: func(cmd *cobra.Command, _ []string) error {
			generator := ctx.App.Generator() // must be declared first so object is constructed on App before calling Start
			defer ctx.App.Stop(cmd.Context())
			err := ctx.App.Start(cmd.Context())
			if err != nil {
				return err
			}

			return generator.Preflight(cmd.Context(), ctx.blockNumber)
		},
	}

	cmd.Flags().StringVarP(&blockNumber, "block-number", "b", "latest", "Block number")

	return cmd
}

func NewPrepareCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputContext{RootContext: rootCtx}
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:     "prepare",
		Short:   "Prepare prover inputs by basing on data previously collected during preflight.",
		Long:    "Prepare prover inputs by basing on data previously collected during preflight. It can be ran off-line in which case it needs --chain-id to be provided",
		PreRunE: preRun(ctx, &blockNumber),
		RunE: func(cmd *cobra.Command, _ []string) error {
			generator := ctx.App.Generator() // must be declared first so object is constructed on App before calling Start
			defer ctx.App.Stop(cmd.Context())
			err := ctx.App.Start(cmd.Context())
			if err != nil {
				return err
			}
			return generator.Prepare(cmd.Context(), ctx.blockNumber)
		},
	}

	cmd.Flags().StringVarP(&blockNumber, "block-number", "b", "latest", "Block number")

	return cmd
}

func NewExecuteCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputContext{RootContext: rootCtx}
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:     "execute",
		Short:   "Execute block by basing on prover inputs previously generated during prepare.",
		Long:    "Execute block by basing on prover inputs previously generated during prepare. It can be ran off-line in which case it needs --chain-id to be provided.",
		PreRunE: preRun(ctx, &blockNumber),
		RunE: func(cmd *cobra.Command, _ []string) error {
			generator := ctx.App.Generator() // must be declared first so object is constructed on App before calling Start
			defer ctx.App.Stop(cmd.Context())
			err := ctx.App.Start(cmd.Context())
			if err != nil {
				return err
			}
			return generator.Execute(cmd.Context(), ctx.blockNumber)
		},
	}

	cmd.Flags().StringVarP(&blockNumber, "block-number", "b", "latest", "Block number")

	return cmd
}

func preRun(ctx *ProverInputContext, blockNumber *string) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if blockNumber != nil {
			var err error
			ctx.blockNumber, err = jsonrpc.FromBlockNumArg(*blockNumber)
			if err != nil {
				return fmt.Errorf("invalid block number: %v", err)
			}
		}

		return nil
	}
}
