package src

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kkrt-labs/go-utils/app"
	"github.com/kkrt-labs/go-utils/common"
	store "github.com/kkrt-labs/go-utils/store"
	compressstore "github.com/kkrt-labs/go-utils/store/compress"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	multistore "github.com/kkrt-labs/go-utils/store/multi"
	s3store "github.com/kkrt-labs/go-utils/store/s3"
	inputstore "github.com/kkrt-labs/zk-pig/src/store"
)

var (
	storeComponentName              = "store"
	fileStoreComponentName          = fmt.Sprintf("%s.file", storeComponentName)
	s3StoreComponentName            = fmt.Sprintf("%s.s3", storeComponentName)
	proverInputStoreComponentName   = "prover-input-store"
	preflightDataStoreComponentName = "preflight-data-store"
)

func (a *App) ProverInputStore() inputstore.ProverInputStore {
	return provide(
		a,
		proverInputStoreComponentName,
		func() (inputstore.ProverInputStore, error) {
			s := a.proverInputStoreBase()
			s = inputstore.ProverInputStoreWithLog(s)
			s = inputstore.ProverInputStoreWithTags(s)

			return s, nil
		},
		app.WithComponentName(proverInputStoreComponentName),
	)
}

func (a *App) proverInputStoreBase() inputstore.ProverInputStore {
	return provide(
		a,
		fmt.Sprintf("%s.base", proverInputStoreComponentName),
		func() (inputstore.ProverInputStore, error) {
			cfg := a.Config().ProverInputs

			return inputstore.NewProverInputStore(a.Store(), common.Val(cfg.ContentType)), nil
		})
}

func (a *App) PreflightDataStore() inputstore.PreflightDataStore {
	return provide(
		a,
		preflightDataStoreComponentName,
		func() (inputstore.PreflightDataStore, error) {
			if a.Config().PreflightData != nil && a.Config().PreflightData.Enabled != nil && *a.Config().PreflightData.Enabled {
				return inputstore.NewPreflightDataStore(a.Store())
			}

			return inputstore.NewNoOpPreflightDataStore(), nil
		},
	)
}

func (a *App) Store() store.Store {
	return provide(
		a,
		storeComponentName,
		func() (store.Store, error) {
			stores := make([]store.Store, 0)
			fileStore := a.FileStore()
			if fileStore != nil {
				stores = append(stores, fileStore)
			}

			s3Store := a.S3Store()
			if s3Store != nil {
				stores = append(stores, s3Store)
			}

			multiStore := multistore.New(stores...)

			compressedStore, err := compressstore.New(multiStore, compressstore.WithContentEncoding(common.Val(a.Config().Store.ContentEncoding)))
			if err != nil {
				return nil, fmt.Errorf("failed to create compressed store: %w", err)
			}

			return compressedStore, nil
		},
	)
}

func (a *App) FileStore() store.Store {
	return provide(
		a,
		fileStoreComponentName,
		func() (store.Store, error) {
			if a.Config().Store.File != nil && a.Config().Store.File.Dir != nil {
				return a.fileStoreWithTags(), nil
			}
			return nil, nil
		},
	)
}

func (a *App) S3Store() store.Store {
	return provide(
		a,
		s3StoreComponentName,
		func() (store.Store, error) {
			if a.Config().Store.S3 != nil && a.Config().Store.S3.Bucket != nil {
				return a.s3StoreWithTags(), nil
			}
			return nil, nil
		},
	)
}

func (a *App) fileStoreBase() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.base", fileStoreComponentName),
		func() (store.Store, error) {
			return filestore.New(common.Val(a.Config().Store.File.Dir)), nil
		},
	)
}

func (a *App) fileStoreWithMetrics() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.metrics", fileStoreComponentName),
		func() (store.Store, error) {
			s := a.fileStoreBase()
			return store.WithMetrics(s), nil
		},
	)
}

func (a *App) fileStoreWithLog() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.logging", fileStoreComponentName),
		func() (store.Store, error) {
			s := a.fileStoreWithMetrics()
			return store.WithLog(s), nil
		},
	)
}

func (a *App) fileStoreWithTags() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.tags", fileStoreComponentName),
		func() (store.Store, error) {
			s := a.fileStoreWithLog()
			return store.WithTags(s), nil
		},
		app.WithComponentName(fileStoreComponentName),
	)
}

func (a *App) s3Client() *s3.Client {
	return provide(
		a,
		fmt.Sprintf("%s.base.s3-client", s3StoreComponentName),
		func() (*s3.Client, error) {
			s3Cfg := a.Config().Store.S3

			opts := make([]func(*awsconfig.LoadOptions) error, 0)

			if s3Cfg.Provider != nil && s3Cfg.Provider.Region != nil {
				opts = append(opts, awsconfig.WithRegion(common.Val(s3Cfg.Provider.Region)))
			}

			if s3Cfg.Provider != nil && s3Cfg.Provider.Credentials != nil && s3Cfg.Provider.Credentials.AccessKey != nil {
				opts = append(opts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					common.Val(s3Cfg.Provider.Credentials.AccessKey),
					common.Val(s3Cfg.Provider.Credentials.SecretKey),
					"",
				)))
			}

			cfg, err := awsconfig.LoadDefaultConfig(
				context.TODO(),
				opts...,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to load default config: %w", err)
			}

			return s3.NewFromConfig(cfg), nil
		},
	)
}

func (a *App) s3StoreBase() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.base", s3StoreComponentName),
		func() (store.Store, error) {
			cfg := a.Config().Store.S3

			return s3store.New(
				a.s3Client(),
				common.Val(cfg.Bucket),
				s3store.WithKeyPrefix(common.Val(cfg.Prefix)),
			)
		},
	)
}

func (a *App) s3StoreWithMetrics() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.metrics", s3StoreComponentName),
		func() (store.Store, error) {
			s := a.s3StoreBase()
			return store.WithMetrics(s), nil
		},
		app.WithComponentName(s3StoreComponentName),
	)
}

func (a *App) s3StoreWithLog() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.logging", s3StoreComponentName),
		func() (store.Store, error) {
			s := a.s3StoreWithMetrics()
			return store.WithLog(s), nil
		},
	)
}

func (a *App) s3StoreWithTags() store.Store {
	return provide(
		a,
		fmt.Sprintf("%s.tags", s3StoreComponentName),
		func() (store.Store, error) {
			s := a.s3StoreWithLog()
			return store.WithTags(s), nil
		},
		app.WithComponentName(s3StoreComponentName),
	)
}
