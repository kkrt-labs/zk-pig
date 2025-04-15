package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	"github.com/kkrt-labs/go-utils/store"
)

//go:generate mockgen -destination=./mock/block_store.go -package=mockstore github.com/kkrt-labs/zk-pig/src/store BlockStore

// BlockStore is a store for blocks.
type BlockStore interface {
	// StoreBlock stores a block.
	StoreBlock(ctx context.Context, chainID uint64, block *ethrpc.Block) error

	// LoadBlock loads a block.
	LoadBlock(ctx context.Context, chainID, blockNumber uint64) (*ethrpc.Block, error)
}

func NewBlockStore(store store.Store) BlockStore {
	return &blockStore{store: store}
}

type blockStore struct {
	store store.Store
}

func (s *blockStore) StoreBlock(ctx context.Context, chainID uint64, block *ethrpc.Block) error {
	path := s.path(chainID, block.Number.ToInt().Uint64())
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(block); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	reader := bytes.NewReader(buf.Bytes())
	headers := store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
		KeyValue: map[string]string{
			"chain.id":     fmt.Sprintf("%d", chainID),
			"block.number": block.Number.ToInt().String(),
		},
	}
	return s.store.Store(ctx, path, reader, &headers)
}

func (s *blockStore) LoadBlock(ctx context.Context, chainID, blockNumber uint64) (*ethrpc.Block, error) {
	path := s.path(chainID, blockNumber)
	block := &ethrpc.Block{}
	reader, _, err := s.store.Load(ctx, path)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(reader).Decode(block); err != nil {
		return nil, err
	}
	return block, nil
}

func (s *blockStore) path(chainID, blockNumber uint64) string {
	return fmt.Sprintf("/%d/blocks/%d.json", chainID, blockNumber)
}

type noOpBlockStore struct{}

func NewNoOpBlockStore() BlockStore {
	return &noOpBlockStore{}
}

func (s *noOpBlockStore) StoreBlock(_ context.Context, _ uint64, _ *ethrpc.Block) error {
	return nil
}

func (s *noOpBlockStore) LoadBlock(_ context.Context, _, _ uint64) (*ethrpc.Block, error) {
	return nil, nil
}
