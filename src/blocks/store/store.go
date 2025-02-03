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
	StoreProverInputs(ctx context.Context, inputs *blockinputs.ProverInputs, headers store.Headers) error

	// LoadProverInputs loads the prover inputs for a block.
	// format can be "protobuf" or "json"
	LoadProverInputs(ctx context.Context, chainID, blockNumber uint64, headers store.Headers) (*blockinputs.ProverInputs, error)
}

type proverInputsStore struct {
	store  store.Store
	format store.ContentType
}

func NewFromStore(inputstore store.Store, format store.ContentType) ProverInputsStore {
	return &proverInputsStore{store: inputstore, format: format}
}

// NewHeavyProverInputsStore creates a new HeavyProverInputsStore instance
func NewHeavyProverInputsStore(cfg *Config) (HeavyProverInputsStore, error) {
	inputstore := filestore.New(filestore.Config{
		DataDir: cfg.MultiConfig.FileConfig.DataDir,
	})

	return &heavyProverInputsStore{
		store:   inputstore,
		baseDir: cfg.MultiConfig.FileConfig.DataDir,
	}, nil
}

type heavyProverInputsStore struct {
	store   store.Store
	baseDir string
}

type Config struct {
	MultiConfig     multistore.Config
	ContentType     store.ContentType
	ContentEncoding store.ContentEncoding
}

func New(cfg *Config) (ProverInputsStore, error) {
	inputstore, err := multistore.New(cfg.MultiConfig)
	if err != nil {
		return nil, err
	}
	return NewFromStore(inputstore, cfg.ContentType), nil
}

func (s *heavyProverInputsStore) StoreHeavyProverInputs(ctx context.Context, inputs *blockinputs.HeavyProverInputs) error {
	path := s.preflightPath(inputs.ChainConfig.ChainID.Uint64(), inputs.Block.Number.ToInt().Uint64())
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(inputs); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	reader := bytes.NewReader(buf.Bytes())
	return s.store.Store(ctx, path, reader, &store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
	})
}

func (s *heavyProverInputsStore) LoadHeavyProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.HeavyProverInputs, error) {
	path := s.preflightPath(chainID, blockNumber)
	data := &blockinputs.HeavyProverInputs{}
	reader, err := s.store.Load(ctx, path, &store.Headers{
		ContentType: store.ContentTypeJSON,
	})
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(reader).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *proverInputsStore) StoreProverInputs(ctx context.Context, data *blockinputs.ProverInputs, headers store.Headers) error {
	var buf bytes.Buffer
	switch s.format {
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
		ct, err := headers.GetContentType()
		if err != nil {
			return fmt.Errorf("failed to get content type: %w", err)
		}
		return fmt.Errorf("unsupported content type: %s", ct)
	}

	path := s.proverPath(data.ChainConfig.ChainID.Uint64(), data.Block.Number.ToInt().Uint64())
	return s.store.Store(ctx, path, bytes.NewReader(buf.Bytes()), &headers)
}

func (s *proverInputsStore) LoadProverInputs(ctx context.Context, chainID, blockNumber uint64, headers store.Headers) (*blockinputs.ProverInputs, error) {
	path := s.proverPath(chainID, blockNumber)
	reader, err := s.store.Load(ctx, path, &headers)
	if err != nil {
		return nil, fmt.Errorf("failed to load data from store: %w", err)
	}

	data := &blockinputs.ProverInputs{}

	switch headers.ContentType {
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
		ct, err := headers.GetContentType()
		if err != nil {
			return nil, fmt.Errorf("failed to get content type: %w", err)
		}
		return nil, fmt.Errorf("unsupported content type: %s", ct)
	}

	return data, nil
}

func (s *heavyProverInputsStore) preflightPath(chainID, blockNumber uint64) string {
	return filepath.Join(fmt.Sprintf("%d", chainID), "preflight", fmt.Sprintf("%d.json", blockNumber))
}

func (s *proverInputsStore) proverPath(chainID, blockNumber uint64) string {
	return filepath.Join(fmt.Sprintf("%d", chainID), fmt.Sprintf("%d", blockNumber))
}
