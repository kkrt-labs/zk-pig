package main

import (
	"context"
	"os"

	ethrpc "github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc/jsonrpc"
	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	"github.com/kkrt-labs/kakarot-controller/src"
	"github.com/kkrt-labs/kakarot-controller/src/blocks"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	cfg := &blocks.Config{
		RPC: &jsonrpchttp.Config{Address: os.Getenv("RPC_URL")},
	}

	logrus.Infof("Version: %s", src.Version)

	svc := blocks.New(cfg)
	err := svc.Generate(context.Background(), ethrpc.MustFromBlockNumArg("latest"))
	if err != nil {
		logrus.Fatalf("Failed to generate block inputs: %v", err)
	}
	logrus.Info("Blocks inputs generated")
}
