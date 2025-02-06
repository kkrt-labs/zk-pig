package blockstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	filestore "github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	multistore "github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
	protoinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs/proto"
	"google.golang.org/protobuf/proto"
)

type BlockStore interface {
	ProverInputsStore
	HeavyProverInputsStore
}

type HeavyProverInputsStore interface {
	// StoreHeavyProverInputs stores the heavy prover inputs for a block.
	StoreHeavyProverInputs(ctx context.Context, inputs *blockinputs.HeavyProverInputs) error

	// LoadHeavyProverInputs loads tthe heavy prover inputs for a block.
	LoadHeavyProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.HeavyProverInputs, error)
}

type ProverInputsStore interface {
	// StoreProverInputs stores the prover inputs for a block.
	StoreProverInputs(ctx context.Context, inputs *blockinputs.ProverInputs) error

	// LoadProverInputs loads the prover inputs for a block.
	// format can be "protobuf" or "json"
	LoadProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.ProverInputs, error)
}

type proverInputsStore struct {
	store       store.Store
	contentType store.ContentType
}

func NewFromStore(inputstore store.Store, contentType store.ContentType) ProverInputsStore {
	return &proverInputsStore{store: inputstore, contentType: contentType}
}

// NewHeavyProverInputsStore creates a new HeavyProverInputsStore instance
func NewHeavyProverInputsStore(cfg *HeavyProverInputsStoreConfig) (HeavyProverInputsStore, error) {
	inputstore := filestore.New(*cfg.FileConfig)

	return &heavyProverInputsStore{
		store: inputstore,
	}, nil
}

type heavyProverInputsStore struct {
	store store.Store
}

type HeavyProverInputsStoreConfig struct {
	FileConfig *filestore.Config
}

type ProverInputsStoreConfig struct {
	MultiStoreConfig multistore.Config
	ContentType      store.ContentType
	ContentEncoding  store.ContentEncoding
}

func New(cfg *ProverInputsStoreConfig) (ProverInputsStore, error) {
	inputstore, err := multistore.NewFromConfig(cfg.MultiStoreConfig)
	if err != nil {
		return nil, err
	}
	return NewFromStore(inputstore, cfg.ContentType), nil
}

func (s *heavyProverInputsStore) StoreHeavyProverInputs(ctx context.Context, inputs *blockinputs.HeavyProverInputs) error {
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

func (s *heavyProverInputsStore) LoadHeavyProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.HeavyProverInputs, error) {
	path := s.preflightPath(blockNumber)
	data := &blockinputs.HeavyProverInputs{}
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

func (s *proverInputsStore) StoreProverInputs(ctx context.Context, data *blockinputs.ProverInputs) error {
	var buf bytes.Buffer
	switch s.contentType {
	case store.ContentTypeProtobuf:
		protoMsg := protoinputs.ToProto(data)
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

func (s *proverInputsStore) LoadProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.ProverInputs, error) {
	path := s.proverPath(blockNumber)
	headers := store.Headers{
		ContentType: s.contentType,
		KeyValue:    map[string]string{"chainID": fmt.Sprintf("%d", chainID)},
	}
	reader, err := s.store.Load(ctx, path, &headers)
	if err != nil {
		return nil, fmt.Errorf("failed to load data from store: %w", err)
	}

	data := &blockinputs.ProverInputs{}

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
		protoMsg := &protoinputs.ProverInputs{}
		if err := proto.Unmarshal(protoBytes, protoMsg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
		}
		data = protoinputs.FromProto(protoMsg)
	default:
		contentType, err := s.contentType.String()
		if err != nil {
			return nil, fmt.Errorf("failed to get content type: %w", err)
		}
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	return data, nil
}

func (s *heavyProverInputsStore) preflightPath(blockNumber uint64) string {
	return filepath.Join(fmt.Sprintf("%d.json", blockNumber))
}

func (s *proverInputsStore) proverPath(blockNumber uint64) string {
	return filepath.Join(fmt.Sprintf("%d", blockNumber))
}
