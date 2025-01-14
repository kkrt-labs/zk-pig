package cmd

import (
	"context"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/jsonrpc"
	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	"github.com/kkrt-labs/kakarot-controller/src/blocks"
	blockstore "github.com/kkrt-labs/kakarot-controller/src/blocks/store"
)

const (
	blockNumberFlag = "block-number"
	formatFlag      = "format"
)

// 1. Main command
func NewProverInputsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prover-inputs",
		Short: "Commands for generating and validating prover inputs",
		RunE:  runHelp,
	}

	cmd.AddCommand(
		NewGenerateCommand(),
		NewPreflightCommand(),
		NewPrepareCommand(),
		NewExecuteCommand(),
	)

	return cmd
}

func runHelp(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

// 2. Subcommands
func NewGenerateCommand() *cobra.Command {
	var (
		rpcURL      string
		dataDir     string
		blockNumber string
		formatStr   string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate prover inputs",
		Long:  "Generate prover inputs by running preflight, prepare and execute in a single run",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := &blocks.Config{
				BaseDir: dataDir,
				RPC:     &jsonrpchttp.Config{Address: rpcURL},
			}

			blockNum, err := parseBigInt(blockNumber, blockNumberFlag)
			if err != nil {
				zap.L().Fatal("Failed to parse block number", zap.Error(err))
			}

			format, err := parseFormat(formatStr)
			if err != nil {
				zap.L().Fatal("Failed to parse format", zap.Error(err))
			}

			svc := blocks.New(cfg)
			if err := svc.Generate(context.Background(), blockNum, format); err != nil {
				zap.L().Fatal("Failed to generate prover inputs", zap.Error(err))
			}
			zap.L().Info("Prover inputs generated")
		},
	}

	addCommonFlags(cmd, &rpcURL, &dataDir, &blockNumber)
	AddFormatFlag(cmd, &formatStr)
	_ = cmd.MarkFlagRequired("rpc-url")

	return cmd
}

func NewPreflightCommand() *cobra.Command {
	var (
		rpcURL      string
		dataDir     string
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:   "preflight",
		Short: "Collect necessary data for proving a block from a remote RPC node",
		Long:  "Collect necessary data for proving a block from a remote RPC node. It processes the EVM block on a state and chain which database have been replaced with a connector to a remote JSON-RPC node",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := &blocks.Config{
				BaseDir: dataDir,
				RPC:     &jsonrpchttp.Config{Address: rpcURL},
			}

			blockNum, err := parseBigInt(blockNumber, "block-number")
			if err != nil {
				zap.L().Fatal("Failed to parse block number", zap.Error(err))
			}

			svc := blocks.New(cfg)
			if err := svc.Preflight(context.Background(), blockNum); err != nil {
				zap.L().Fatal("Preflight failed", zap.Error(err))
			}
			zap.L().Info("Preflight succeeded")
		},
	}

	addCommonFlags(cmd, &rpcURL, &dataDir, &blockNumber)
	_ = cmd.MarkFlagRequired("rpc-url")

	return cmd
}

func NewPrepareCommand() *cobra.Command {
	var (
		rpcURL      string
		dataDir     string
		blockNumber string
		chainID     string
		formatStr   string
	)

	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepare prover inputs, basing on data collected during preflight",
		Long:  "Prepare prover inputs, basing on data collected during preflight. It processes and validates an EVM block over in memory state and chain prefilled with data collected during preflight.",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := &blocks.Config{
				BaseDir: dataDir,
				RPC:     &jsonrpchttp.Config{Address: rpcURL},
			}

			blockNum, err := parseBigInt(blockNumber, "block-number")
			if err != nil {
				zap.L().Fatal("Failed to parse block number", zap.Error(err))
			}
			chainIDBig, err := parseBigInt(chainID, "chain-id")
			if err != nil {
				zap.L().Fatal("Failed to parse chain-id", zap.Error(err))
			}

			format, err := parseFormat(formatStr)
			if err != nil {
				zap.L().Fatal("Failed to parse format", zap.Error(err))
			}

			svc := blocks.New(cfg)
			if err := svc.Prepare(context.Background(), chainIDBig, blockNum, format); err != nil {
				zap.L().Fatal("Failed to prepare prover inputs", zap.Error(err))
			}
			zap.L().Info("Prover inputs prepared")
		},
	}

	addCommonFlags(cmd, &rpcURL, &dataDir, &blockNumber)
	AddFormatFlag(cmd, &formatStr)
	cmd.Flags().StringVar(&chainID, "chain-id", "", "Chain ID (decimal)")
	_ = cmd.MarkFlagRequired("chain-id")

	return cmd
}

func NewExecuteCommand() *cobra.Command {
	var (
		rpcURL      string
		dataDir     string
		blockNumber string
		chainID     string
		formatStr   string
	)

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Run an EVM execution, basing on prover inputs generated during prepare",
		Long:  "Run an EVM execution, basing on prover inputs generated during prepare. It processes and validates an EVM block over in memory state and chain prefilled with prover inputs.",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := &blocks.Config{
				BaseDir: dataDir,
				RPC:     &jsonrpchttp.Config{Address: rpcURL},
			}

			blockNum, err := parseBigInt(blockNumber, "block-number")
			if err != nil {
				zap.L().Fatal("Failed to parse block number", zap.Error(err))
			}
			chainIDBig, err := parseBigInt(chainID, "chain-id")
			if err != nil {
				zap.L().Fatal("Failed to parse chain-id", zap.Error(err))
			}

			format, err := parseFormat(formatStr)
			if err != nil {
				zap.L().Fatal("Failed to parse format", zap.Error(err))
			}

			svc := blocks.New(cfg)
			if err := svc.Execute(context.Background(), chainIDBig, blockNum, format); err != nil {
				zap.L().Fatal("Execute failed", zap.Error(err))
			}
			zap.L().Info("Execute succeeded")
		},
	}

	addCommonFlags(cmd, &rpcURL, &dataDir, &blockNumber)
	AddFormatFlag(cmd, &formatStr)
	cmd.Flags().StringVar(&chainID, "chain-id", "", "Chain ID (decimal)")
	_ = cmd.MarkFlagRequired("chain-id")

	return cmd
}

// 3. Common flag helpers
func addCommonFlags(cmd *cobra.Command, rpcURL, dataDir, blockNumber *string) {
	f := cmd.Flags()
	rpcURLFlag := AddRPCURLFlag(rpcURL, f)
	AddDataDirFlag(dataDir, f)
	blockNumberFlag := AddBlockNumberFlag(blockNumber, f)

	_ = cmd.MarkFlagRequired(blockNumberFlag)
	_ = cmd.MarkFlagRequired(rpcURLFlag)
}

func AddRPCURLFlag(rpcURL *string, f *pflag.FlagSet) string {
	flagName := "rpc-url"
	f.StringVar(rpcURL, flagName, "", "Chain JSON-RPC URL")
	return flagName
}

func AddDataDirFlag(dataDir *string, f *pflag.FlagSet) string {
	flagName := "data-dir"
	f.StringVar(dataDir, flagName, "data", "Path to data directory")
	return flagName
}

func AddBlockNumberFlag(blockNumber *string, f *pflag.FlagSet) string {
	flagName := blockNumberFlag
	f.StringVar(blockNumber, flagName, "", "Block number")
	return flagName
}

func AddChainIDFlag(chainID *string, f *pflag.FlagSet) string {
	flagName := "chain-id"
	f.StringVar(chainID, flagName, "", "Chain ID (decimal)")
	return flagName
}

func AddFormatFlag(cmd *cobra.Command, format *string) {
	cmd.Flags().StringVar(format, formatFlag, "json", "Output format (json or protobuf)")
}

func parseBigInt(val, flagName string) (*big.Int, error) {
	if val == "" {
		return nil, fmt.Errorf("missing required flag: %s", flagName)
	}

	// Use FromBlockNumArg for block-number flag to support special values
	if flagName == blockNumberFlag {
		bn, err := jsonrpc.FromBlockNumArg(val)
		if err != nil {
			return nil, fmt.Errorf("invalid block number %q: %w", val, err)
		}
		return bn, nil
	}

	// For other flags (like chain-id), keep using decimal only
	bn := new(big.Int)
	if _, ok := bn.SetString(val, 10); !ok {
		return nil, fmt.Errorf("invalid integer value %q for flag %s", val, flagName)
	}
	return bn, nil
}

func parseFormat(formatStr string) (blockstore.Format, error) {
	if formatStr == "" || formatStr == "json" {
		return blockstore.JSONFormat, nil
	} else if formatStr == "protobuf" {
		return blockstore.ProtobufFormat, nil
	}
	return 0, fmt.Errorf("invalid format %q", formatStr)
}
