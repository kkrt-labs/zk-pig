package multistore

import (
	"context"
	"fmt"
	"io"

	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	filestore "github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	s3store "github.com/kkrt-labs/kakarot-controller/pkg/store/s3"
)

type Store struct {
	stores []store.Store
}

func NewFromConfig(cfg Config) (store.Store, error) {
	var stores []store.Store

	if cfg.FileConfig != nil {
		stores = append(stores, filestore.New(*cfg.FileConfig))
	}

	if cfg.S3Config != nil {
		s3Store, err := s3store.New(cfg.S3Config)
		if err != nil {
			return nil, err
		}
		stores = append(stores, s3Store)
	}

	return &Store{stores: stores}, nil
}

func New(stores []store.Store) store.Store {
	return &Store{stores: stores}
}

func (m *Store) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	for _, s := range m.stores {
		if err := s.Store(ctx, key, reader, headers); err != nil {
			return err
		}
	}
	return nil
}

func (m *Store) Load(ctx context.Context, key string, headers *store.Headers) (io.Reader, error) {
	// Try stores in order until we find the data or encounter an error
	for _, s := range m.stores {
		reader, err := s.Load(ctx, key, headers)
		if err != nil {
			return nil, fmt.Errorf("failed to load from store: %w", err)
		}
		if reader != nil {
			return reader, nil
		}
	}
	return nil, fmt.Errorf("key %s not found in any store", key)
}
