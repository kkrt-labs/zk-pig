package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	store "github.com/kkrt-labs/go-utils/store"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
)

type HeavyProverInputStore interface {
	// StoreHeavyProverInput stores the heavy prover inputs for a block.
	StoreHeavyProverInput(ctx context.Context, inputs *input.HeavyProverInput) error

	// LoadHeavyProverInput loads tthe heavy prover inputs for a block.
	LoadHeavyProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.HeavyProverInput, error)
}

// NewHeavyProverInputStore creates a new HeavyProverInputStore instance
func NewHeavyProverInputStore(cfg *HeavyProverInputStoreConfig) (HeavyProverInputStore, error) {
	inputstore := filestore.New(*cfg.FileConfig)

	return &heavyProverInputtore{
		store: inputstore,
	}, nil
}

type heavyProverInputtore struct {
	store store.Store
}

type HeavyProverInputStoreConfig struct {
	FileConfig *filestore.Config
}

func (s *heavyProverInputtore) StoreHeavyProverInput(ctx context.Context, inputs *input.HeavyProverInput) error {
	path := s.preflightPath(inputs.Block.Number.ToInt().Uint64())
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(inputs); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	reader := bytes.NewReader(buf.Bytes())
	headers := store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
		KeyValue:        map[string]string{"chainID": fmt.Sprintf("%d", inputs.ChainConfig.ChainID.Uint64())},
	}
	return s.store.Store(ctx, path, reader, &headers)
}

func (s *heavyProverInputtore) LoadHeavyProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.HeavyProverInput, error) {
	path := s.preflightPath(blockNumber)
	data := &input.HeavyProverInput{}
	headers := store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
		KeyValue:        map[string]string{"chainID": fmt.Sprintf("%d", chainID)},
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

func (s *heavyProverInputtore) preflightPath(blockNumber uint64) string {
	return fmt.Sprintf("%d.json", blockNumber)
}
