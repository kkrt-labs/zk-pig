package main

import (
	"context"
	"fmt"
	"os"

	ethrpc "github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/jsonrpc"
	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	"github.com/kkrt-labs/kakarot-controller/src"
	"github.com/kkrt-labs/kakarot-controller/src/blocks"
	"go.uber.org/zap"
)

func main() {
	level := os.Getenv("LOGGER_LEVEL")
	var logger *zap.Logger
	switch level {
	case "debug":
		logger, _ = zap.NewDevelopment(zap.IncreaseLevel(zap.DebugLevel))
	case "info":
		logger, _ = zap.NewDevelopment(zap.IncreaseLevel(zap.InfoLevel))
	default:
		logger, _ = zap.NewDevelopment(zap.IncreaseLevel(zap.InfoLevel))
	}
	zap.ReplaceGlobals(logger)
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()

	cfg := &blocks.Config{
		RPC: &jsonrpchttp.Config{Address: os.Getenv("RPC_URL")},
	}

	logger.Info("Version", zap.String("version", src.Version))

	svc := blocks.New(cfg)
	err := svc.Generate(context.Background(), ethrpc.MustFromBlockNumArg("latest"))
	if err != nil {
		logger.Fatal("Failed to generate block inputs", zap.Error(err))
	}
	logger.Info("Blocks inputs generated")
}
