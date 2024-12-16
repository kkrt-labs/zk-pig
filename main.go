package main

import (
	"context"
	"os"

	ethrpc "github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/jsonrpc"
	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	"github.com/kkrt-labs/kakarot-controller/src"
	"github.com/kkrt-labs/kakarot-controller/src/blocks"
	"go.uber.org/zap"
)

func main() {
	// TODO: configure dev/prod environments to use zap.NewProduction() in production
	// and zap.NewDevelopment() in dev. We can also modify log levels (debug, info, etc.)
	logger, _ := zap.NewDevelopment(zap.IncreaseLevel(zap.InfoLevel))
	defer logger.Sync() // flushes buffer, if any

	cfg := &blocks.Config{
		RPC: &jsonrpchttp.Config{Address: os.Getenv("RPC_URL")},
	}

	logrus.Infof("Version: %s", src.Version)

	svc := blocks.New(cfg)
	err := svc.Generate(context.Background(), ethrpc.MustFromBlockNumArg("latest"))
	if err != nil {
		logger.Fatal("Failed to generate block inputs", zap.Error(err))
	}
	logger.Info("Blocks inputs generated")
}
