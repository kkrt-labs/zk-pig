package src

import (
	"context"
	"fmt"
	"math/big"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cenkalti/backoff/v4"
	"github.com/kkrt-labs/go-utils/app"
	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	ethjsonrpc "github.com/kkrt-labs/go-utils/ethereum/rpc/jsonrpc"
	"github.com/kkrt-labs/go-utils/jsonrpc"
	jsonrpcmrgd "github.com/kkrt-labs/go-utils/jsonrpc/merged"
	store "github.com/kkrt-labs/go-utils/store"
	compressstore "github.com/kkrt-labs/go-utils/store/compress"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	multistore "github.com/kkrt-labs/go-utils/store/multi"
	s3store "github.com/kkrt-labs/go-utils/store/s3"
	"github.com/kkrt-labs/zk-pig/src/ethereum/evm"
	"github.com/kkrt-labs/zk-pig/src/generator"
	"github.com/kkrt-labs/zk-pig/src/steps"
	inputstore "github.com/kkrt-labs/zk-pig/src/store"
	"go.uber.org/zap"
)

var (
	preflightDataStoreComponentName     = "preflight-data-store"
	preflightDataFileStoreComponentName = fmt.Sprintf("%s.file", preflightDataStoreComponentName)
	proverInputStoreComponentName       = "prover-input-store"
	proverInputFileStoreComponentName   = fmt.Sprintf("%s.file", proverInputStoreComponentName)
	proverInputS3StoreComponentName     = fmt.Sprintf("%s.s3", proverInputStoreComponentName)
	chainComponentName                  = "chain"
	chainRPCComponentName               = fmt.Sprintf("%s.rpc", chainComponentName)
	generatorComponentName              = "generator"
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
		app.WithName("zkpig"),
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

func provide[T any](a *App, name string, constructor func() (T, error), opts ...app.ServiceOption) T {
	return app.Provide(a.app, name, constructor, opts...)
}

func (a *App) ChainRPCBase() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.base", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
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
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) ChainRPCMetrics() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.metrics", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
			remote := a.ChainRPCBase()
			if remote == nil {
				return nil, nil
			}

			remote = jsonrpc.WithMetrics(remote)

			return remote, nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) ChainRPCLogging() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.logging", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
			remote := a.ChainRPCSecured()
			if remote == nil {
				return nil, nil
			}

			remote = jsonrpc.WithLog()(remote)

			return remote, nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) ChainRPCSecured() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.secured", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
			remote := a.ChainRPCMetrics()
			if remote == nil {
				return nil, nil
			}

			remote = jsonrpc.WithTimeout(500 * time.Millisecond)(remote)
			remote = jsonrpc.WithExponentialBackOffRetry(
				backoff.WithInitialInterval(50*time.Millisecond),
				backoff.WithMaxElapsedTime(2*time.Second),
			)(remote)

			return remote, nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) ChainRPCTagged() jsonrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.tagged", chainRPCComponentName),
		func() (jsonrpc.Client, error) {
			remote := a.ChainRPCLogging()
			if remote == nil {
				return nil, nil
			}

			return jsonrpc.WithTags(remote), nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) ChainRPC() jsonrpc.Client {
	return provide(
		a,
		chainRPCComponentName,
		func() (jsonrpc.Client, error) {
			remote := a.ChainRPCTagged()
			if remote == nil {
				return nil, nil
			}

			remote = jsonrpc.WithVersion("2.0")(remote)
			remote = jsonrpc.WithIncrementalID()(remote)

			return remote, nil
		},
		app.WithComponentName(chainRPCComponentName),
	)
}

func (a *App) ChainBase() ethrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.base", chainComponentName),
		func() (ethrpc.Client, error) {
			remote := a.ChainRPC()
			if remote == nil {
				return nil, nil
			}
			return ethjsonrpc.NewFromClient(remote), nil
		},
		app.WithComponentName(chainComponentName),
	)
}

func (a *App) ChainWithMetrics() ethrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.metrics", chainComponentName),
		func() (ethrpc.Client, error) {
			chain := a.ChainBase()
			if chain == nil {
				return nil, nil
			}
			return ethrpc.WithMetrics(chain), nil
		},
		app.WithComponentName(chainComponentName),
	)
}

func (a *App) Chain() ethrpc.Client {
	return provide(
		a,
		fmt.Sprintf("%s.check", chainComponentName),
		func() (ethrpc.Client, error) {
			chain := a.ChainWithMetrics()
			if chain == nil {
				return nil, nil
			}
			return ethrpc.WithCheck(chain), nil
		},
		app.WithComponentName(chainComponentName),
	)
}

func (a *App) ChainID() *big.Int {
	return provide(
		a,
		fmt.Sprintf("%s.id", chainComponentName),
		func() (*big.Int, error) {
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

func (a *App) PreflightDataFileStoreBase() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.base", preflightDataFileStoreComponentName),
		func() (store.Store, error) {
			gCfg := a.Config()
			if gCfg.PreflightDataStore.File.Dir != "" {
				return filestore.New(filepath.Join(gCfg.DataDir, gCfg.PreflightDataStore.File.Dir)), nil
			}
			return nil, nil
		},
	)
}

func (a *App) PreflightDataFileStoreWithMetrics() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.metrics", preflightDataFileStoreComponentName),
		func() (store.Store, error) {
			s := a.PreflightDataFileStoreBase()
			if s == nil {
				return nil, nil
			}
			return store.WithMetrics(s), nil
		},
	)
}

func (a *App) PreflightDataFileStoreWithLog() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.logging", preflightDataFileStoreComponentName),
		func() (store.Store, error) {
			s := a.PreflightDataFileStoreWithMetrics()
			if s == nil {
				return nil, nil
			}
			return store.WithLog(s), nil
		},
	)
}

func (a *App) PreflightDataStoreBase() inputstore.PreflightDataStore {
	return provide(
		a,
		fmt.Sprintf("%s.base", preflightDataStoreComponentName),
		func() (inputstore.PreflightDataStore, error) {
			s := a.PreflightDataFileStoreWithLog()
			if s != nil {
				return inputstore.NewPreflightDataStore(s)
			}

			return nil, nil
		},
	)
}

func (a *App) PreflightDataStore() inputstore.PreflightDataStore {
	return provide(
		a,
		preflightDataStoreComponentName,
		func() (inputstore.PreflightDataStore, error) {
			s := a.PreflightDataStoreBase()
			s = inputstore.PreflightDataStoreWithLog(s)
			s = inputstore.PreflightDataStoreWithTags(s)
			return s, nil
		},
		app.WithComponentName(preflightDataStoreComponentName),
	)
}
func (a *App) ProverInputFileStoreBase() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.base", proverInputFileStoreComponentName),
		func() (store.Store, error) {
			gCfg := a.Config()
			if gCfg.ProverInputStore.File.Dir != "" {
				return filestore.New(filepath.Join(gCfg.DataDir, gCfg.ProverInputStore.File.Dir)), nil
			}
			return nil, nil
		},
	)
}

func (a *App) ProverInputFileStoreWithMetrics() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.metrics", proverInputFileStoreComponentName),
		func() (store.Store, error) {
			s := a.ProverInputFileStoreBase()
			if s == nil {
				return nil, nil
			}
			return store.WithMetrics(s), nil
		},
		app.WithComponentName(proverInputFileStoreComponentName),
	)
}

func (a *App) ProverInputFileStoreWithLog() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.logging", proverInputFileStoreComponentName),
		func() (store.Store, error) {
			s := a.ProverInputFileStoreWithMetrics()
			if s == nil {
				return nil, nil
			}
			return store.WithLog(s), nil
		},
	)
}

func (a *App) ProverInputFileStoreWithTags() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.tags", proverInputFileStoreComponentName),
		func() (store.Store, error) {
			s := a.ProverInputFileStoreWithLog()
			if s == nil {
				return nil, nil
			}
			return store.WithTags(s), nil
		},
		app.WithComponentName(proverInputFileStoreComponentName),
	)
}

func (a *App) ProverInputS3StoreBaseS3Client() *s3.Client {
	return provide(
		a,
		fmt.Sprintf("%s.base.s3-client", proverInputS3StoreComponentName),
		func() (*s3.Client, error) {
			gCfg := a.Config()
			if gCfg.ProverInputStore.S3.AWSProvider.Region != "" {
				return s3.NewFromConfig(aws.Config{
					Region: gCfg.ProverInputStore.S3.AWSProvider.Region,
					Credentials: credentials.NewStaticCredentialsProvider(
						gCfg.ProverInputStore.S3.AWSProvider.Credentials.AccessKey,
						gCfg.ProverInputStore.S3.AWSProvider.Credentials.SecretKey,
						"",
					),
				}), nil
			}
			return nil, nil
		},
	)
}

func (a *App) ProverInputS3StoreBase() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.base", proverInputS3StoreComponentName),
		func() (store.Store, error) {
			gCfg := a.Config()
			if gCfg.ProverInputStore.S3.Bucket != "" {
				return s3store.New(
					a.ProverInputS3StoreBaseS3Client(),
					gCfg.ProverInputStore.S3.Bucket,
					s3store.WithKeyPrefix(gCfg.ProverInputStore.S3.BucketKeyPrefix),
				)
			}
			return nil, nil
		},
	)
}

func (a *App) ProverInputS3StoreWithMetrics() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.metrics", proverInputS3StoreComponentName),
		func() (store.Store, error) {
			s := a.ProverInputS3StoreBase()
			if s == nil {
				return nil, nil
			}
			return store.WithMetrics(s), nil
		},
		app.WithComponentName(proverInputS3StoreComponentName),
	)
}

func (a *App) ProverInputS3StoreWithLog() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.logging", proverInputS3StoreComponentName),
		func() (store.Store, error) {
			s := a.ProverInputS3StoreWithMetrics()
			if s == nil {
				return nil, nil
			}
			return store.WithLog(s), nil
		},
	)
}

func (a *App) ProverInputS3StoreWithTags() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.tags", proverInputS3StoreComponentName),
		func() (store.Store, error) {
			s := a.ProverInputS3StoreWithLog()
			if s == nil {
				return nil, nil
			}
			return store.WithTags(s), nil
		},
		app.WithComponentName(proverInputS3StoreComponentName),
	)
}

func (a *App) ProverInputStoreBase() inputstore.ProverInputStore {
	return provide(
		a,
		fmt.Sprintf("%s.base", proverInputStoreComponentName),
		func() (inputstore.ProverInputStore, error) {
			gCfg := a.Config()

			contentType, err := store.ParseContentType(gCfg.ProverInputStore.ContentType)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ProverInputStore content type: %v", err)
			}

			stores := make([]store.Store, 0)
			fileStore := a.ProverInputFileStoreWithTags()
			if fileStore != nil {
				stores = append(stores, fileStore)
			}

			s3Store := a.ProverInputS3StoreWithTags()
			if s3Store != nil {
				stores = append(stores, s3Store)
			}

			multiStore := multistore.New(stores...)

			contentEncoding, err := store.ParseContentEncoding(gCfg.ProverInputStore.ContentEncoding)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ProverInputStore content encoding: %v", err)
			}

			compressedStore, err := compressstore.New(multiStore, compressstore.WithContentEncoding(contentEncoding))
			if err != nil {
				return nil, fmt.Errorf("failed to create compressed store: %v", err)
			}

			return inputstore.NewProverInputStore(compressedStore, contentType), nil
		})
}

func (a *App) ProverInputStore() inputstore.ProverInputStore {
	return provide(
		a,
		proverInputStoreComponentName,
		func() (inputstore.ProverInputStore, error) {
			s := a.ProverInputStoreBase()
			s = inputstore.ProverInputStoreWithLog(s)
			s = inputstore.ProverInputStoreWithTags(s)

			return s, nil
		},
		app.WithComponentName(proverInputStoreComponentName),
	)
}
func (a *App) PreflightEVM() evm.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.preflight.evm", generatorComponentName),
		func() (evm.Executor, error) {
			vm := evm.NewExecutor()
			vm = evm.WithLog()(vm)
			vm = evm.WithTags(vm)
			return vm, nil
		},
	)
}

func (a *App) PreflightBase() steps.Preflight {
	return provide(
		a,
		fmt.Sprintf("%s.preflight.base", generatorComponentName),
		func() (steps.Preflight, error) {
			return steps.NewPreflightFromEvm(a.PreflightEVM(), a.Chain()), nil
		},
	)
}

func (a *App) Preflight() steps.Preflight {
	return provide(
		a,
		fmt.Sprintf("%s.preflight", generatorComponentName),
		func() (steps.Preflight, error) {
			return steps.PreflightWithTags(a.PreflightBase()), nil
		},
	)
}

func (a *App) PreparerEVM() evm.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.preparer.evm", generatorComponentName),
		func() (evm.Executor, error) {
			vm := evm.NewExecutor()
			vm = evm.WithLog()(vm)
			vm = evm.WithTags(vm)
			return vm, nil
		},
	)
}

func (a *App) PreparerBase() steps.Preparer {
	return provide(
		a,
		fmt.Sprintf("%s.preparer.base", generatorComponentName),
		func() (steps.Preparer, error) {
			gCfg := a.Config()
			inclusion, err := steps.ParseIncludes(gCfg.Generator.Include...)
			if err != nil {
				return nil, nil
			}
			return steps.NewPreparerFromEvm(a.PreparerEVM(), steps.WithDataInclude(inclusion))
		},
	)
}

func (a *App) Preparer() steps.Preparer {
	return provide(
		a,
		fmt.Sprintf("%s.preparer", generatorComponentName),
		func() (steps.Preparer, error) {
			return steps.PreparerWithTags(a.PreparerBase()), nil
		},
	)
}

func (a *App) ExecutorEVM() evm.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.executor.evm", generatorComponentName),
		func() (evm.Executor, error) {
			vm := evm.NewExecutor()
			vm = evm.WithLog()(vm)
			vm = evm.WithTags(vm)
			return vm, nil
		},
	)
}

func (a *App) ExecutorBase() steps.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.executor.base", generatorComponentName),
		func() (steps.Executor, error) {
			return steps.NewExecutorFromEvm(a.ExecutorEVM()), nil
		},
	)
}

func (a *App) Executor() steps.Executor {
	return provide(
		a,
		fmt.Sprintf("%s.executor", generatorComponentName),
		func() (steps.Executor, error) {
			return steps.ExecutorWithTags(a.ExecutorBase()), nil
		},
	)
}

func (a *App) Generator() *generator.Generator {
	return provide(
		a,
		fmt.Sprintf("%s.base", generatorComponentName),
		func() (*generator.Generator, error) {
			return generator.NewGenerator(
				&generator.Config{
					ChainID:            a.ChainID(),
					RPC:                a.Chain(),
					Preflighter:        a.Preflight(),
					Preparer:           a.Preparer(),
					Executor:           a.Executor(),
					PreflightDataStore: a.PreflightDataStore(),
					ProverInputStore:   a.ProverInputStore(),
				},
			)
		},
		app.WithComponentName(generatorComponentName), // override component name
	)
}

func (a *App) Daemon() *generator.Daemon {
	return provide(
		a,
		fmt.Sprintf("%s.daemon", generatorComponentName),
		func() (*generator.Daemon, error) {
			a.app.EnableHealthzEntrypoint()
			return &generator.Daemon{
				Generator: a.Generator(),
			}, nil
		},
		app.WithComponentName(generatorComponentName), // override component name
	)
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
