package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	store "github.com/kkrt-labs/go-utils/store"
	"github.com/kkrt-labs/zk-pig/src/steps"
)

//go:generate mockgen -destination=./mock/preflight_data_store.go -package=mockstore github.com/kkrt-labs/zk-pig/src/store PreflightDataStore

// PreflightDataStore is a store for preflight data.
type PreflightDataStore interface {
	// StorePreflightData stores preflight data for a block.
	StorePreflightData(ctx context.Context, inputs *steps.PreflightData) error

	// LoadPreflightData loads preflight data inputs for a block.
	LoadPreflightData(ctx context.Context, chainID, blockNumber uint64) (*steps.PreflightData, error)
}

// NewPreflightDataStore creates a new PreflightDataStore instance
func NewPreflightDataStore(store store.Store) (PreflightDataStore, error) {
	return &preflightDataStore{
		store: store,
	}, nil
}

type preflightDataStore struct {
	store store.Store
}

func (s *preflightDataStore) StorePreflightData(ctx context.Context, data *steps.PreflightData) error {
	chainID := data.ChainConfig.ChainID.Uint64()
	blockNumber := data.Block.Number.ToInt().Uint64()
	path := s.path(chainID, blockNumber)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	reader := bytes.NewReader(buf.Bytes())
	headers := store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
		KeyValue: map[string]string{
			"chain.id":     fmt.Sprintf("%d", chainID),
			"block.number": fmt.Sprintf("%d", blockNumber),
		},
	}
	return s.store.Store(ctx, path, reader, &headers)
}

func (s *preflightDataStore) LoadPreflightData(ctx context.Context, chainID, blockNumber uint64) (*steps.PreflightData, error) {
	path := s.path(chainID, blockNumber)
	data := &steps.PreflightData{}
	reader, _, err := s.store.Load(ctx, path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	if err := json.NewDecoder(reader).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *preflightDataStore) path(chainID, blockNumber uint64) string {
	return fmt.Sprintf("/%d/%d/preflight.json", chainID, blockNumber)
}

type noOpPreflightDataStore struct{}

func (s *noOpPreflightDataStore) StorePreflightData(_ context.Context, _ *steps.PreflightData) error {
	return nil
}

func (s *noOpPreflightDataStore) LoadPreflightData(_ context.Context, _, _ uint64) (*steps.PreflightData, error) {
	return nil, nil
}

func NewNoOpPreflightDataStore() PreflightDataStore {
	return &noOpPreflightDataStore{}
}
