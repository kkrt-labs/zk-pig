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
	"github.com/kkrt-labs/zk-pig/pkg/app"
	"github.com/kkrt-labs/zk-pig/src/generator"
	"github.com/kkrt-labs/zk-pig/src/steps"
	inputstore "github.com/kkrt-labs/zk-pig/src/store"
	"go.uber.org/zap"
)

// Service is a service that enables the generation of prover inpunts for EVM compatible blocks.
type App struct {
	app *app.App
	cfg *Config
}

func NewApp(cfg *Config, logger *zap.Logger) (*App, error) {
	a, err := app.NewApp(
		&cfg.App,
		app.WithLogger(logger),
		app.WithName("zk-pig"),
		app.WithVersion(Version),
	)
	if err != nil {
		return nil, err
	}

	return &App{
		app: a,
		cfg: cfg,
	}, nil
}

func (a *App) Config() *Config {
	return a.cfg
}

func (a *App) BaseJSONRPC() jsonrpc.Client {
	return app.Provide(a.app, "base-jsonrpc", func() (jsonrpc.Client, error) {
		gCfg := a.Config()
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

func (a *App) JSONRPC() jsonrpc.Client {
	return app.Provide(a.app, "jsonrpc", func() (jsonrpc.Client, error) {
		remote := a.BaseJSONRPC()
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

func (a *App) EthRPC() *ethjsonrpc.Client {
	return app.Provide(a.app, "ethrpc", func() (*ethjsonrpc.Client, error) {
		jrpc := a.JSONRPC()
		if jrpc == nil {
			return nil, fmt.Errorf("jsonrpc client not set")
		}
		return ethjsonrpc.NewFromClient(jrpc), nil
	})
}

func (a *App) ChainID() *big.Int {
	return app.Provide(a.app, "chain-id", func() (*big.Int, error) {
		gCfg := a.Config()
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

func (a *App) PreflightDataStore() inputstore.PreflightDataStore {
	return app.Provide(a.app, "preflight-data-store", func() (inputstore.PreflightDataStore, error) {
		gCfg := a.Config()
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

func (a *App) ProverInputStore() inputstore.ProverInputStore {
	return app.Provide(a.app, "prover-input-store", func() (inputstore.ProverInputStore, error) {
		gCfg := a.Config()

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

func (a *App) Preflight() steps.Preflight {
	return app.Provide(a.app, "preflight", func() (steps.Preflight, error) {
		return steps.NewPreflight(a.EthRPC()), nil
	})
}

func (a *App) Preparer() steps.Preparer {
	return app.Provide(a.app, "preparer", func() (steps.Preparer, error) {
		gCfg := a.Config()
		inclusion, err := steps.ParseInclude(gCfg.Generator.Include...)
		if err != nil {
			return nil, nil
		}
		return steps.NewPreparer(
			steps.WithInclusion(inclusion),
		)
	})
}

func (a *App) Executor() steps.Executor {
	return app.Provide(a.app, "executor", func() (steps.Executor, error) {
		return steps.NewExecutor(), nil
	})
}

func (a *App) Generator() *generator.Generator {
	return app.Provide(a.app, "generator", func() (*generator.Generator, error) {
		return &generator.Generator{
			ChainID:            a.ChainID(),
			RPC:                a.EthRPC(),
			Preflighter:        a.Preflight(),
			Preparer:           a.Preparer(),
			Executor:           a.Executor(),
			PreflightDataStore: a.PreflightDataStore(),
			ProverInputStore:   a.ProverInputStore(),
		}, nil
	})
}

func (a *App) Daemon() *generator.Daemon {
	return app.Provide(a.app, "daemon", func() (*generator.Daemon, error) {
		a.app.EnableMain()
		a.app.EnableHealthz()
		return &generator.Daemon{
			Generator: a.Generator(),
		}, nil
	})
}

func (a *App) Start(ctx context.Context) error {
	return a.app.Start(ctx)
}

func (a *App) Stop(ctx context.Context) error {
	return a.app.Stop(ctx)
}

func (a *App) Run(ctx context.Context) error {
	return a.app.Run(ctx)
}

func (a *App) Error() error {
	return a.app.Error()
}
