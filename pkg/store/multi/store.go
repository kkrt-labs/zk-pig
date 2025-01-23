package multi

import (
	"context"
	"fmt"
	"io"

	"github.com/kkrt-labs/kakarot-controller/pkg/store"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/s3"
)

type multiStore struct {
	stores []store.Store
}

func New(cfg Config) (store.Store, error) {
	var stores []store.Store

	if cfg.FileConfig != nil {
		stores = append(stores, file.New(*cfg.FileConfig))
	}

	if cfg.S3Config != nil {
		s3Store, err := s3.New(cfg.S3Config)
		if err != nil {
			return nil, err
		}
		stores = append(stores, s3Store)
	}

	return &multiStore{stores: stores}, nil
}

func (m *multiStore) Store(ctx context.Context, key string, reader io.Reader, headers *store.Headers) error {
	for _, s := range m.stores {
		if err := s.Store(ctx, key, reader, headers); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiStore) Load(ctx context.Context, key string, headers *store.Headers) (io.Reader, error) {
	// Try stores in order until we find the data
	for _, s := range m.stores {
		if reader, err := s.Load(ctx, key, headers); err == nil {
			return reader, nil
		}
	}
	return nil, fmt.Errorf("key %s not found in any store", key)
}
