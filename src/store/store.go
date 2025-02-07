package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	store "github.com/kkrt-labs/go-utils/store"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	multistore "github.com/kkrt-labs/go-utils/store/multi"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	protoinput "github.com/kkrt-labs/zk-pig/src/prover-input/proto"
	"google.golang.org/protobuf/proto"
)

type BlockStore interface {
	ProverInputStore
	HeavyProverInputStore
}

type HeavyProverInputStore interface {
	// StoreHeavyProverInput stores the heavy prover inputs for a block.
	StoreHeavyProverInput(ctx context.Context, inputs *input.HeavyProverInput) error

	// LoadHeavyProverInput loads tthe heavy prover inputs for a block.
	LoadHeavyProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.HeavyProverInput, error)
}

type ProverInputStore interface {
	// StoreProverInput stores the prover inputs for a block.
	StoreProverInput(ctx context.Context, inputs *input.ProverInput) error

	// LoadProverInput loads the prover inputs for a block.
	// format can be "protobuf" or "json"
	LoadProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.ProverInput, error)
}

type proverInputsStore struct {
	store       store.Store
	contentType store.ContentType
}

func NewFromStore(inputstore store.Store, contentType store.ContentType) ProverInputStore {
	return &proverInputsStore{store: inputstore, contentType: contentType}
}

// NewHeavyProverInputStore creates a new HeavyProverInputStore instance
func NewHeavyProverInputStore(cfg *HeavyProverInputStoreConfig) (HeavyProverInputStore, error) {
	inputstore := filestore.New(*cfg.FileConfig)

	return &heavyProverInputStore{
		store: inputstore,
	}, nil
}

type heavyProverInputStore struct {
	store store.Store
}

type HeavyProverInputStoreConfig struct {
	FileConfig *filestore.Config
}

type ProverInputStoreConfig struct {
	MultiStoreConfig multistore.Config
	ContentType      store.ContentType
	ContentEncoding  store.ContentEncoding
}

func New(cfg *ProverInputStoreConfig) (ProverInputStore, error) {
	inputstore, err := multistore.NewFromConfig(cfg.MultiStoreConfig)
	if err != nil {
		return nil, err
	}
	return NewFromStore(inputstore, cfg.ContentType), nil
}

func (s *heavyProverInputStore) StoreHeavyProverInput(ctx context.Context, inputs *input.HeavyProverInput) error {
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

func (s *heavyProverInputStore) LoadHeavyProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.HeavyProverInput, error) {
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

func (s *proverInputsStore) StoreProverInput(ctx context.Context, data *input.ProverInput) error {
	var buf bytes.Buffer
	switch s.contentType {
	case store.ContentTypeProtobuf:
		protoMsg := protoinput.ToProto(data)
		protoBytes, err := proto.Marshal(protoMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf: %w", err)
		}
		buf.Write(protoBytes)
	case store.ContentTypeJSON:
		if err := json.NewEncoder(&buf).Encode(data); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	default:
		contentType, err := s.contentType.String()
		if err != nil {
			return fmt.Errorf("failed to get content type: %w", err)
		}
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	path := s.proverPath(data.Block.Number.ToInt().Uint64())
	headers := store.Headers{
		ContentType: s.contentType,
		KeyValue:    map[string]string{"chainID": fmt.Sprintf("%d", data.ChainConfig.ChainID.Uint64())},
	}
	return s.store.Store(ctx, path, bytes.NewReader(buf.Bytes()), &headers)
}

func (s *proverInputsStore) LoadProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.ProverInput, error) {
	path := s.proverPath(blockNumber)
	headers := store.Headers{
		ContentType: s.contentType,
		KeyValue:    map[string]string{"chainID": fmt.Sprintf("%d", chainID)},
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
		contentType, err := s.contentType.String()
		if err != nil {
			return nil, fmt.Errorf("failed to get content type: %w", err)
		}
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	return data, nil
}

func (s *heavyProverInputStore) preflightPath(blockNumber uint64) string {
	return fmt.Sprintf("%d.json", blockNumber)
}

func (s *proverInputsStore) proverPath(blockNumber uint64) string {
	return fmt.Sprintf("%d", blockNumber)
}
