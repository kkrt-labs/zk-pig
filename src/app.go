package src

import (
	"context"
	"fmt"
	"math/big"
	"path/filepath"
	"time"

	aws "github.com/kkrt-labs/go-utils/aws"
	ethjsonrpc "github.com/kkrt-labs/go-utils/ethereum/rpc/jsonrpc"
	"github.com/kkrt-labs/go-utils/jsonrpc"
	jsonrpcmrgd "github.com/kkrt-labs/go-utils/jsonrpc/merged"
	store "github.com/kkrt-labs/go-utils/store"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	multistore "github.com/kkrt-labs/go-utils/store/multi"
	s3store "github.com/kkrt-labs/go-utils/store/s3"
	"github.com/kkrt-labs/zk-pig/pkg/svc"
	"github.com/kkrt-labs/zk-pig/src/generator"
	"github.com/kkrt-labs/zk-pig/src/steps"
	inputstore "github.com/kkrt-labs/zk-pig/src/store"
	"go.uber.org/zap"
)

// Service is a service that enables the generation of prover inpunts for EVM compatible blocks.
type App struct {
	svc.App
	cfg *Config
}

func NewApp(cfg *Config, logger *zap.Logger) (*App, error) {
	return &App{
		App: *svc.NewApp(logger),
		cfg: cfg,
	}, nil
}

func (app *App) Config() *Config {
	return app.cfg
}

func (app *App) BaseJSONRPC() jsonrpc.Client {
	return svc.Provide(&(app.App), "base-jsonrpc", func() (jsonrpc.Client, error) {
		gCfg := app.Config()
		if gCfg.Chain.RPC.URL != "" {
			cfg := (&jsonrpcmrgd.Config{
				Addr: gCfg.Chain.RPC.URL,
			}).SetDefault()

			remote, err := jsonrpcmrgd.New(cfg)
			if err != nil {
				return nil, err
			}

			return remote, nil
		}
		return nil, nil
	})
}

func (app *App) JSONRPC() jsonrpc.Client {
	return svc.Provide(&(app.App), "jsonrpc", func() (jsonrpc.Client, error) {
		remote := app.BaseJSONRPC()
		if remote == nil {
			return nil, nil
		}

		remote = jsonrpc.WithLog()(remote)                           // Logs a first time before the Retry
		remote = jsonrpc.WithTimeout(500 * time.Millisecond)(remote) // Sets a timeout on outgoing requests
		remote = jsonrpc.WithTags("")(remote)                        // Add tags are updated according to retry
		remote = jsonrpc.WithRetry()(remote)
		remote = jsonrpc.WithTags("jsonrpc")(remote)
		remote = jsonrpc.WithVersion("2.0")(remote)
		remote = jsonrpc.WithIncrementalID()(remote)

		return remote, nil
	})
}

func (app *App) EthRPC() *ethjsonrpc.Client {
	return svc.Provide(&(app.App), "ethrpc", func() (*ethjsonrpc.Client, error) {
		jrpc := app.JSONRPC()
		if jrpc == nil {
			return nil, fmt.Errorf("jsonrpc client not set")
		}
		return ethjsonrpc.NewFromClient(jrpc), nil
	})
}

func (app *App) ChainID() *big.Int {
	return svc.Provide(&(app.App), "chain-id", func() (*big.Int, error) {
		gCfg := app.Config()
		if gCfg.Chain.ID != "" {
			chainID, ok := new(big.Int).SetString(gCfg.Chain.ID, 10)
			if !ok {
				return nil, fmt.Errorf("failed to parse chain ID: %s", gCfg.Chain.ID)
			}
			return chainID, nil
		}

		return nil, nil
	})
}

func (app *App) PreflightDataStore() inputstore.PreflightDataStore {
	return svc.Provide(&(app.App), "preflight-data-store", func() (inputstore.PreflightDataStore, error) {
		gCfg := app.Config()
		if gCfg.PreflightDataStore.File.Dir != "" {
			cfg := &inputstore.PreflightDataStoreConfig{
				FileConfig: &filestore.Config{
					DataDir: filepath.Join(gCfg.DataDir, gCfg.PreflightDataStore.File.Dir),
				},
			}
			return inputstore.NewPreflightDataStore(cfg)
		}

		return nil, nil
	})
}

func (app *App) ProverInputStore() inputstore.ProverInputStore {
	return svc.Provide(&(app.App), "prover-input-store", func() (inputstore.ProverInputStore, error) {
		gCfg := app.Config()

		var proverInputStoreCfg multistore.Config

		// If File Dir config is set file store
		if gCfg.ProverInputStore.File.Dir != "" {
			proverInputStoreCfg.FileConfig = &filestore.Config{DataDir: filepath.Join(gCfg.DataDir, gCfg.ProverInputStore.File.Dir)}
		}

		// Configure S3 store
		if gCfg.ProverInputStore.S3.Bucket != "" {
			proverInputStoreCfg.S3Config = &s3store.Config{
				Bucket:    gCfg.ProverInputStore.S3.Bucket,
				KeyPrefix: gCfg.ProverInputStore.S3.BucketKeyPrefix,
				ProviderConfig: &aws.ProviderConfig{
					Region: gCfg.ProverInputStore.S3.AWSProvider.Region,
					Credentials: &aws.CredentialsConfig{
						AccessKey: gCfg.ProverInputStore.S3.AWSProvider.Credentials.AccessKey,
						SecretKey: gCfg.ProverInputStore.S3.AWSProvider.Credentials.SecretKey,
					},
				},
			}
		}

		contentEncoding, err := store.ParseContentEncoding(gCfg.ProverInputStore.ContentEncoding)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content encoding: %v", err)
		}
		contentType, err := store.ParseContentType(gCfg.ProverInputStore.ContentType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content type: %v", err)
		}

		// Set prover inputs store configuration
		cfg := &inputstore.ProverInputStoreConfig{
			StoreConfig:     proverInputStoreCfg,
			ContentEncoding: contentEncoding,
			ContentType:     contentType,
		}

		return inputstore.New(cfg)
	})
}

func (app *App) Preflight() steps.Preflight {
	return svc.Provide(&(app.App), "preflight", func() (steps.Preflight, error) {
		return steps.NewPreflight(app.EthRPC()), nil
	})
}

func (app *App) Preparer() steps.Preparer {
	return svc.Provide(&(app.App), "preparer", func() (steps.Preparer, error) {
		return steps.NewPreparer(), nil
	})
}

func (app *App) Executor() steps.Executor {
	return svc.Provide(&(app.App), "executor", func() (steps.Executor, error) {
		return steps.NewExecutor(), nil
	})
}

func (app *App) Generator() *generator.Generator {
	return svc.Provide(&(app.App), "generator", func() (*generator.Generator, error) {
		return &generator.Generator{
			ChainID:            app.ChainID(),
			RPC:                app.EthRPC(),
			Preflighter:        app.Preflight(),
			Preparer:           app.Preparer(),
			Executor:           app.Executor(),
			PreflightDataStore: app.PreflightDataStore(),
			ProverInputStore:   app.ProverInputStore(),
		}, nil
	})
}

func (app *App) Daemon() *generator.Daemon {
	return svc.Provide(&(app.App), "daemon", func() (*generator.Daemon, error) {
		return &generator.Daemon{
			Generator: app.Generator(),
		}, nil
	})
}

func (app *App) Start(ctx context.Context) error {
	return app.App.Start(ctx)
}

func (app *App) Stop(ctx context.Context) error {
	return app.App.Stop(ctx)
}
