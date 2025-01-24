package cmd

import (
	"fmt"
	"math/big"

	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/jsonrpc"
	filestore "github.com/kkrt-labs/kakarot-controller/pkg/store"
	"github.com/kkrt-labs/kakarot-controller/src/blocks"
	"github.com/kkrt-labs/kakarot-controller/src/config"
	"github.com/spf13/cobra"
)

type ProverInputsContext struct {
	RootContext
	svc         *blocks.Service
	blockNumber *big.Int
	format      filestore.ContentType
	compression filestore.ContentEncoding
	storage     string
	s3Bucket    string
	keyPrefix   string
	accessKey   string
	secretKey   string
}

// 1. Main command
func NewProverInputsCommand(rootCtx *RootContext) *cobra.Command {
	var (
		ctx         = &ProverInputsContext{RootContext: *rootCtx}
		blockNumber string
		format      string
		compression string
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

			ctx.format, err = filestore.ParseFormat(format)
			if err != nil {
				return fmt.Errorf("invalid format: %v", err)
			}

			ctx.compression, err = filestore.ParseCompression(compression)
			if err != nil {
				return fmt.Errorf("invalid compression: %v", err)
			}

			// Validate storage type
			if ctx.storage != "file" && ctx.storage != "s3" {
				return fmt.Errorf("invalid storage type: %s (must be 'file' or 's3')", ctx.storage)
			}

			// Set default keyPrefix based on storage type
			if ctx.storage == "file" && ctx.keyPrefix == "" {
				ctx.keyPrefix = "./data"
			}

			if ctx.storage == "s3" {
				if ctx.s3Bucket == "" {
					return fmt.Errorf("s3-bucket must be specified when using s3 storage")
				}
				if ctx.keyPrefix == "" {
					return fmt.Errorf("key-prefix must be specified when using s3 storage")
				}
				if ctx.accessKey == "" {
					return fmt.Errorf("access-key must be specified when using s3 storage")
				}
				if ctx.secretKey == "" {
					return fmt.Errorf("secret-key must be specified when using s3 storage")
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
	cmd.PersistentFlags().StringVarP(&format, "format", "f", "json", fmt.Sprintf("Format for storing prover inputs (one of %q)", []string{"json", "protobuf"}))
	cmd.PersistentFlags().StringVarP(&compression, "compression", "z", "none", fmt.Sprintf("Compression for storing prover inputs (one of %q)", []string{"none", "flate", "zlib"}))
	cmd.PersistentFlags().StringVar(&ctx.storage, "storage", "file", "Storage type (file or s3)")
	cmd.PersistentFlags().StringVar(&ctx.s3Bucket, "s3-bucket", "", "S3 bucket name for storing prover inputs")
	cmd.PersistentFlags().StringVar(&ctx.keyPrefix, "key-prefix", "", "Key prefix for storing prover inputs")
	cmd.PersistentFlags().StringVar(&ctx.accessKey, "access-key", "", "Access key for storing prover inputs")
	cmd.PersistentFlags().StringVar(&ctx.secretKey, "secret-key", "", "Secret key for storing prover inputs")

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
			headers := filestore.Headers{
				ContentType:     ctx.format,
				ContentEncoding: ctx.compression,
				KeyValue:        map[string]string{"storage": ctx.storage, "s3-bucket": ctx.s3Bucket, "key-prefix": ctx.keyPrefix, "access-key": ctx.accessKey, "secret-key": ctx.secretKey},
			}
			return ctx.svc.Generate(cmd.Context(), ctx.blockNumber, headers)
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
			headers := filestore.Headers{
				ContentType:     ctx.format,
				ContentEncoding: ctx.compression,
				KeyValue:        map[string]string{"storage": ctx.storage, "s3-bucket": ctx.s3Bucket, "key-prefix": ctx.keyPrefix, "access-key": ctx.accessKey, "secret-key": ctx.secretKey},
			}
			return ctx.svc.Prepare(cmd.Context(), ctx.blockNumber, headers)
		},
	}
}

func NewExecuteCommand(ctx *ProverInputsContext) *cobra.Command {
	return &cobra.Command{
		Use:   "execute",
		Short: "Run an EVM execution, basing on prover inputs generated during prepare",
		Long:  "Run an EVM execution, basing on prover inputs generated during prepare. It processes and validates an EVM block over in memory state and chain prefilled with prover inputs.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			headers := filestore.Headers{
				ContentType:     ctx.format,
				ContentEncoding: ctx.compression,
				KeyValue:        map[string]string{"storage": ctx.storage, "s3-bucket": ctx.s3Bucket, "key-prefix": ctx.keyPrefix, "access-key": ctx.accessKey, "secret-key": ctx.secretKey},
			}
			return ctx.svc.Execute(cmd.Context(), ctx.blockNumber, headers)
		},
	}
}
