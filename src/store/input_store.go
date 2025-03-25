package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	store "github.com/kkrt-labs/go-utils/store"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	protoinput "github.com/kkrt-labs/zk-pig/src/prover-input/proto"
	"google.golang.org/protobuf/proto"
)

//go:generate mockgen -destination=./mock/input_store.go -package=mockstore github.com/kkrt-labs/zk-pig/src/store ProverInputStore

// ProverInputStore is a store for prover inputs.
type ProverInputStore interface {
	// StoreProverInput stores the prover inputs for a block.
	StoreProverInput(ctx context.Context, inputs *input.ProverInput) error

	// LoadProverInput loads the prover inputs for a block.
	// format can be "protobuf" or "json"
	LoadProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.ProverInput, error)
}

type proverInputStore struct {
	store       store.Store
	contentType store.ContentType
}

func NewProverInputStore(s store.Store, contentType store.ContentType) ProverInputStore {
	return &proverInputStore{store: s, contentType: contentType}
}

func (s *proverInputStore) StoreProverInput(ctx context.Context, data *input.ProverInput) error {
	buf := new(bytes.Buffer)
	switch s.contentType {
	case store.ContentTypeProtobuf:
		protoMsg := protoinput.ToProto(data)
		protoBytes, err := proto.Marshal(protoMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf: %w", err)
		}
		buf.Write(protoBytes)
	case store.ContentTypeJSON:
		if err := json.NewEncoder(buf).Encode(data); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported content type: %s", s.contentType)
	}

	path := s.path(data.ChainConfig.ChainID.Uint64(), data.Blocks[0].Header.Number.Uint64())
	headers := &store.Headers{
		ContentType: s.contentType,
	}
	return s.store.Store(ctx, path, bytes.NewReader(buf.Bytes()), headers)
}

func (s *proverInputStore) LoadProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.ProverInput, error) {
	path := s.path(chainID, blockNumber)
	headers := store.Headers{
		ContentType: s.contentType,
	}
	reader, err := s.store.Load(ctx, path, &headers)
	if err != nil {
		return nil, fmt.Errorf("failed to load data from store: %w", err)
	}

	data := &input.ProverInput{}

	switch s.contentType {
	case store.ContentTypeJSON:
		if err := json.NewDecoder(reader).Decode(data); err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}
	case store.ContentTypeProtobuf:
		protoBytes, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read protobuf data: %w", err)
		}
		protoMsg := &protoinput.ProverInput{}
		if err := proto.Unmarshal(protoBytes, protoMsg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
		}
		data = protoinput.FromProto(protoMsg)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", s.contentType)
	}

	return data, nil
}

func (s *proverInputStore) path(chainID, blockNumber uint64) string {
	return fmt.Sprintf("%d/%d", chainID, blockNumber)
}
