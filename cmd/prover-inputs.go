package cmd

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	"github.com/kkrt-labs/kakarot-controller/src/blocks"
)

// Global flags for all subcommands
var (
	blockNumber string
	chainID     string
	rpcURL      string
	dataDir     string
	logLevel    string
	logFormat   string
)

var proverInputsCmd = &cobra.Command{
	Use:   "prover-inputs",
	Short: "Commands for managing prover inputs",
	Long: strings.TrimSpace(`
Commands to handle the entire pipeline of proving EVM blocks:
  - generate: Runs preflight + prepare + execute
  - preflight: Only fetch and store preflight data
  - prepare: Uses the heavy input from preflight to create final ProverInputs
  - execute: Verifies the final ProverInputs by re-executing the block
`),
}

// generateCmd: runs preflight + prepare + execute
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate prover inputs (preflight + prepare + execute)",
	Run: func(_ *cobra.Command, _ []string) {
		setupLogger()

		cfg := &blocks.Config{
			BaseDir: dataDir,
			RPC:     &jsonrpchttp.Config{Address: rpcURL},
		}

		blockNum := parseBigIntOrDie(blockNumber, "block-number")

		svc := blocks.New(cfg)
		if err := svc.Generate(context.Background(), blockNum); err != nil {
			zap.L().Fatal("Failed to generate block inputs", zap.Error(err))
		}
		zap.L().Info("Blocks inputs generated")
	},
}

// preflightCmd: only the "preflight" step
var preflightCmd = &cobra.Command{
	Use:   "preflight",
	Short: "Run preflight checks",
	Run: func(_ *cobra.Command, _ []string) {
		setupLogger()

		cfg := &blocks.Config{
			BaseDir: dataDir,
			RPC:     &jsonrpchttp.Config{Address: rpcURL},
		}

		blockNum := parseBigIntOrDie(blockNumber, "block-number")

		svc := blocks.New(cfg)
		if err := svc.Preflight(context.Background(), blockNum); err != nil {
			zap.L().Fatal("Preflight failed", zap.Error(err))
		}
		zap.L().Info("Preflight checks complete")
	},
}

// prepareCmd: uses heavy input to create final ProverInputs
var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare prover inputs",
	Run: func(_ *cobra.Command, _ []string) {
		setupLogger()

		cfg := &blocks.Config{
			BaseDir: dataDir,
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

// executeCmd: verifies final ProverInputs by re-executing the block
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute prover inputs generation",
	Run: func(_ *cobra.Command, _ []string) {
		setupLogger()

		cfg := &blocks.Config{
			BaseDir: dataDir,
			RPC:     &jsonrpchttp.Config{Address: rpcURL},
		}

		blockNum := parseBigIntOrDie(blockNumber, "block-number")
		chainIDBig := parseBigIntOrDie(chainID, "chain-id")

		svc := blocks.New(cfg)
		if err := svc.Execute(context.Background(), chainIDBig, blockNum); err != nil {
			zap.L().Fatal("Failed to execute prover inputs generation", zap.Error(err))
		}
		zap.L().Info("Prover inputs generation executed")
	},
}

func init() {
	proverInputsCmd.AddCommand(
		generateCmd,
		preflightCmd,
		prepareCmd,
		executeCmd,
	)

	for _, c := range []*cobra.Command{generateCmd, preflightCmd, prepareCmd, executeCmd} {
		c.Flags().StringVar(&blockNumber, "block-number", os.Getenv("BLOCK_NUMBER"), "Block number (decimal)")
		c.Flags().StringVar(&rpcURL, "rpc-url", os.Getenv("RPC_URL"), "Ethereum RPC URL (default from RPC_URL env var)")
		c.Flags().StringVar(&dataDir, "data-dir", os.Getenv("DATA_DIR"), "Path to data directory (default from DATA_DIR env var)")
		c.Flags().StringVar(&logLevel, "log-level", os.Getenv("LOG_LEVEL"), "Log level (debug|info|warn|error)")
		c.Flags().StringVar(&logFormat, "log-format", os.Getenv("LOG_FORMAT"), "Log format (json|text)")

		_ = c.MarkFlagRequired("block-number")
		_ = c.MarkFlagRequired("rpc-url")
		_ = c.MarkFlagRequired("data-dir")
	}

	for _, c := range []*cobra.Command{prepareCmd, executeCmd} {
		c.Flags().StringVar(&chainID, "chain-id", "", "Chain ID (decimal)")
		_ = c.MarkFlagRequired("chain-id")
	}
}

// parseBigIntOrDie converts a string to *big.Int or logs fatal
func parseBigIntOrDie(val, flagName string) *big.Int {
	if val == "" {
		zap.L().Fatal("Missing required flag", zap.String("flag", flagName))
	}
	bn := new(big.Int)
	if _, ok := bn.SetString(val, 10); !ok {
		zap.L().Fatal("Invalid integer value", zap.String("flag", flagName), zap.String("value", val))
	}
	return bn
}

// setupLogger configures zap based on --log-level and --log-format
func setupLogger() {
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
		zap.L().Warn("Unknown log-level, using default info", zap.String("log-level", logLevel))
	}

	// Log Format
	if strings.EqualFold(logFormat, "text") {
		cfg.Encoding = "console"
	} else {
		cfg.Encoding = "json"
	}

	logger, err := cfg.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build logger: %v\n", err)
		os.Exit(1)
	}
	zap.ReplaceGlobals(logger)
}
