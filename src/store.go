package src

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kkrt-labs/go-utils/app"
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

			contentType, err := store.ParseContentType(cfg.ContentType)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ProverInputStore content type: %v", err)
			}

			return inputstore.NewProverInputStore(a.Store(), contentType), nil
		})
}

func (a *App) PreflightDataStore() inputstore.PreflightDataStore {
	return provide(
		a,
		preflightDataStoreComponentName,
		func() (inputstore.PreflightDataStore, error) {
			if a.Config().PreflightData.Enabled {
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

			contentEncoding, err := store.ParseContentEncoding(a.Config().Store.ContentEncoding)
			if err != nil {
				return nil, fmt.Errorf("failed to parse content encoding: %w", err)
			}

			compressedStore, err := compressstore.New(multiStore, compressstore.WithContentEncoding(contentEncoding))
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
			if a.Config().Store.File.Dir != "" {
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
			if a.Config().Store.S3.Bucket != "" {
				return a.fileStoreWithTags(), nil
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
			return filestore.New(a.Config().Store.File.Dir), nil
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
			cfg := a.Config().Store.S3
			return s3.NewFromConfig(aws.Config{
				Region: cfg.AWSProvider.Region,
				Credentials: credentials.NewStaticCredentialsProvider(
					cfg.AWSProvider.Credentials.AccessKey,
					cfg.AWSProvider.Credentials.SecretKey,
					"",
				),
			}), nil
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
				cfg.Bucket,
				s3store.WithKeyPrefix(cfg.Prefix),
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
