package cmd

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/jsonrpc"
	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	"github.com/kkrt-labs/kakarot-controller/src/blocks"
)

const (
	blockNumberFlag = "block-number"
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
		logLevel    string
		logFormat   string
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate prover inputs",
		Long:  "Generate prover inputs by running preflight, prepare and execute in a single run",
		Run: func(_ *cobra.Command, _ []string) {
			if err := setupLogger(logLevel, logFormat); err != nil {
				zap.L().Fatal("Failed to setup logger", zap.Error(err))
			}

			cfg := &blocks.Config{
				BaseDir: defaultIfEmpty(dataDir),
				RPC:     &jsonrpchttp.Config{Address: rpcURL},
			}

			blockNum := parseBigIntOrDie(blockNumber, "block-number")

			svc := blocks.New(cfg)
			if err := svc.Generate(context.Background(), blockNum); err != nil {
				zap.L().Fatal("Failed to generate prover inputs", zap.Error(err))
			}
			zap.L().Info("Prover inputs generated")
		},
	}

	addCommonFlags(cmd, &rpcURL, &dataDir, &logLevel, &logFormat, &blockNumber)
	_ = cmd.MarkFlagRequired("rpc-url")

	return cmd
}

func NewPreflightCommand() *cobra.Command {
	var (
		rpcURL      string
		dataDir     string
		logLevel    string
		logFormat   string
		blockNumber string
	)

	cmd := &cobra.Command{
		Use:   "preflight",
		Short: "Collect necessary data for proving a block from a remote RPC node",
		Long:  "Collect necessary data for proving a block from a remote RPC node. It processes the EVM block on a state and chain which database have been replaced with a connector to a remote JSON-RPC node",
		Run: func(_ *cobra.Command, _ []string) {
			if err := setupLogger(logLevel, logFormat); err != nil {
				zap.L().Fatal("Failed to setup logger", zap.Error(err))
			}

			cfg := &blocks.Config{
				BaseDir: defaultIfEmpty(dataDir),
				RPC:     &jsonrpchttp.Config{Address: rpcURL},
			}

			blockNum := parseBigIntOrDie(blockNumber, "block-number")

			svc := blocks.New(cfg)
			if err := svc.Preflight(context.Background(), blockNum); err != nil {
				zap.L().Fatal("Preflight failed", zap.Error(err))
			}
			zap.L().Info("Preflight succeeded")
		},
	}

	addCommonFlags(cmd, &rpcURL, &dataDir, &logLevel, &logFormat, &blockNumber)
	_ = cmd.MarkFlagRequired("rpc-url")

	return cmd
}

func NewPrepareCommand() *cobra.Command {
	var (
		rpcURL      string
		dataDir     string
		logLevel    string
		logFormat   string
		blockNumber string
		chainID     string
	)

	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepare prover inputs, basing on data collected during preflight",
		Long:  "Prepare prover inputs, basing on data collected during preflight. It processes and validates an EVM block over in memory state and chain prefilled with data collected during preflight.",
		Run: func(_ *cobra.Command, _ []string) {
			if err := setupLogger(logLevel, logFormat); err != nil {
				zap.L().Fatal("Failed to setup logger", zap.Error(err))
			}

			cfg := &blocks.Config{
				BaseDir: defaultIfEmpty(dataDir),
				RPC:     &jsonrpchttp.Config{Address: rpcURL},
			}

			blockNum := parseBigIntOrDie(blockNumber, "block-number")
			chainIDBig := parseBigIntOrDie(chainID, "chain-id")

			svc := blocks.New(cfg)
			if err := svc.Prepare(context.Background(), chainIDBig, blockNum); err != nil {
				zap.L().Fatal("Failed to prepare prover inputs", zap.Error(err))
			}
			zap.L().Info("Prover inputs prepared")
		},
	}

	addCommonFlags(cmd, &rpcURL, &dataDir, &logLevel, &logFormat, &blockNumber)
	cmd.Flags().StringVar(&chainID, "chain-id", "", "Chain ID (decimal)")
	_ = cmd.MarkFlagRequired("chain-id")

	return cmd
}

func NewExecuteCommand() *cobra.Command {
	var (
		rpcURL      string
		dataDir     string
		logLevel    string
		logFormat   string
		blockNumber string
		chainID     string
	)

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Run an EVM execution, basing on prover inputs generated during prepare",
		Long:  "Run an EVM execution, basing on prover inputs generated during prepare. It processes and validates an EVM block over in memory state and chain prefilled with prover inputs.",
		Run: func(_ *cobra.Command, _ []string) {
			if err := setupLogger(logLevel, logFormat); err != nil {
				zap.L().Fatal("Failed to setup logger", zap.Error(err))
			}

			cfg := &blocks.Config{
				BaseDir: defaultIfEmpty(dataDir),
				RPC:     &jsonrpchttp.Config{Address: rpcURL},
			}

			blockNum := parseBigIntOrDie(blockNumber, "block-number")
			chainIDBig := parseBigIntOrDie(chainID, "chain-id")

			svc := blocks.New(cfg)
			if err := svc.Execute(context.Background(), chainIDBig, blockNum); err != nil {
				zap.L().Fatal("Execute failed", zap.Error(err))
			}
			zap.L().Info("Execute succeeded")
		},
	}

	addCommonFlags(cmd, &rpcURL, &dataDir, &logLevel, &logFormat, &blockNumber)
	cmd.Flags().StringVar(&chainID, "chain-id", "", "Chain ID (decimal)")
	_ = cmd.MarkFlagRequired("chain-id")

	return cmd
}

// 3. Common flag helpers
func addCommonFlags(cmd *cobra.Command, rpcURL, dataDir, logLevel, logFormat, blockNumber *string) {
	f := cmd.Flags()
	rpcURLFlag := AddRPCURLFlag(rpcURL, f)
	AddDataDirFlag(dataDir, f)
	AddLogLevelFlag(logLevel, f)
	AddLogFormatFlag(logFormat, f)
	blockNumberFlag := AddBlockNumberFlag(blockNumber, f)

	_ = cmd.MarkFlagRequired(blockNumberFlag)
	_ = cmd.MarkFlagRequired(rpcURLFlag)
}

func AddRPCURLFlag(rpcURL *string, f *pflag.FlagSet) string {
	flagName := "rpc-url"
	f.StringVar(rpcURL, flagName, "RPC_URL", "Ethereum RPC URL (default from RPC_URL env var)")
	return flagName
}

func AddDataDirFlag(dataDir *string, f *pflag.FlagSet) string {
	flagName := "data-dir"
	f.StringVar(dataDir, flagName, "data", "Path to data directory")
	return flagName
}

func AddLogLevelFlag(logLevel *string, f *pflag.FlagSet) string {
	flagName := "log-level"
	f.StringVar(logLevel, flagName, "info", "Log level (debug|info|warn|error)")
	return flagName
}

func AddLogFormatFlag(logFormat *string, f *pflag.FlagSet) string {
	flagName := "log-format"
	f.StringVar(logFormat, flagName, "text", "Log format (json|text)")
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

// 4. Utility functions
func setupLogger(logLevel, logFormat string) error {
	cfg := zap.NewProductionConfig()

	// Log Level
	switch strings.ToLower(logLevel) {
	case "debug":
		cfg.Level.SetLevel(zap.DebugLevel)
	case "info":
		cfg.Level.SetLevel(zap.InfoLevel)
	case "warn":
		cfg.Level.SetLevel(zap.WarnLevel)
	case "error":
		cfg.Level.SetLevel(zap.ErrorLevel)
	case "":
		// do nothing, keep default from Production
	default:
		return fmt.Errorf("invalid log-level %q, must be one of: debug, info, warn, error", logLevel)
	}

	// Log Format
	if strings.EqualFold(logFormat, "text") {
		cfg.Encoding = "console"
	} else {
		cfg.Encoding = "json"
	}

	logger, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}
	zap.ReplaceGlobals(logger)
	return nil
}

func parseBigIntOrDie(val, flagName string) *big.Int {
	if val == "" {
		zap.L().Fatal("Missing required flag", zap.String("flag", flagName))
	}

	// Use FromBlockNumArg for block-number flag to support special values
	if flagName == blockNumberFlag {
		bn, err := jsonrpc.FromBlockNumArg(val)
		if err != nil {
			zap.L().Fatal("Invalid block number",
				zap.String("flag", flagName),
				zap.String("value", val),
				zap.Error(err))
		}
		return bn
	}

	// For other flags (like chain-id), keep using decimal only
	bn := new(big.Int)
	if _, ok := bn.SetString(val, 10); !ok {
		zap.L().Fatal("Invalid integer value",
			zap.String("flag", flagName),
			zap.String("value", val))
	}
	return bn
}

func defaultIfEmpty(value string) string {
	if value == "" {
		return "data"
	}
	return value
}
