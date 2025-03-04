package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	store "github.com/kkrt-labs/go-utils/store"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	"github.com/kkrt-labs/zk-pig/src/steps"
)

type PreflightDataStore interface {
	// StorePreflightData stores preflight data for a block.
	StorePreflightData(ctx context.Context, inputs *steps.PreflightData) error

	// LoadPreflightData loads preflight data inputs for a block.
	LoadPreflightData(ctx context.Context, chainID, blockNumber uint64) (*steps.PreflightData, error)
}

// NewPreflightDataStore creates a new PreflightDataStore instance
func NewPreflightDataStore(cfg *PreflightDataStoreConfig) (PreflightDataStore, error) {
	inputstore := filestore.New(*cfg.FileConfig)

	return &preflightDataStore{
		store: inputstore,
	}, nil
}

type preflightDataStore struct {
	store store.Store
}

type PreflightDataStoreConfig struct {
	FileConfig *filestore.Config
}

func (s *preflightDataStore) StorePreflightData(ctx context.Context, data *steps.PreflightData) error {
	path := s.preflightPath(data.ChainConfig.ChainID.Uint64(), data.Block.Number.ToInt().Uint64())
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	reader := bytes.NewReader(buf.Bytes())
	headers := store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
	}
	return s.store.Store(ctx, path, reader, &headers)
}

func (s *preflightDataStore) LoadPreflightData(ctx context.Context, chainID, blockNumber uint64) (*steps.PreflightData, error) {
	path := s.preflightPath(chainID, blockNumber)
	data := &steps.PreflightData{}
	headers := store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
	}
	reader, err := s.store.Load(ctx, path, &headers)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(reader).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *preflightDataStore) preflightPath(chainID, blockNumber uint64) string {
	return fmt.Sprintf("%d/%d.json", chainID, blockNumber)
}
