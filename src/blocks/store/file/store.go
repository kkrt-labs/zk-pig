package fileblockstore

// Implementation of BlockStore interface that stores the preflight and prover inputs in files.
//
// The preflight data is stored in at path `<base-dir>/<chainID>/preflight/<blockNumber>.json`
// The prover inputs are stored in a file named `<base-dir>/<chainID>/prover/<blockNumber>.json`

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/kkrt-labs/kakarot-controller/pkg/aws"
	storeinputs "github.com/kkrt-labs/kakarot-controller/pkg/store"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/compress"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/s3"
	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
	protoinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs/proto"
	"google.golang.org/protobuf/proto"
)

type ProverInputsStore struct {
	store   storeinputs.Store
	format  storeinputs.ContentType
	baseDir string
}

func NewFromStore(store storeinputs.Store, format storeinputs.ContentType) *ProverInputsStore {
	return &ProverInputsStore{store: store, format: format}
}

type HeavyProverInputsStore struct {
	store   storeinputs.Store
	baseDir string
}

type Config struct {
	MultiConfig     multi.Config
	ContentType     storeinputs.ContentType
	ContentEncoding storeinputs.ContentEncoding
}

func New(cfg *Config) (*ProverInputsStore, error) {
	store, err := multi.New(cfg.MultiConfig)
	if err != nil {
		return nil, err
	}
	return NewFromStore(store, cfg.ContentType), nil
}

func (s *HeavyProverInputsStore) StoreHeavyProverInputs(ctx context.Context, inputs *blockinputs.HeavyProverInputs) error {
	path := s.preflightPath(inputs.ChainConfig.ChainID.Uint64(), inputs.Block.Number.ToInt().Uint64())
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(inputs); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	reader := bytes.NewReader(buf.Bytes())
	return s.store.Store(ctx, path, reader, &storeinputs.Headers{
		ContentType:     storeinputs.ContentTypeJSON,
		ContentEncoding: storeinputs.ContentEncodingPlain,
	})
}

func (s *HeavyProverInputsStore) LoadHeavyProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.HeavyProverInputs, error) {
	path := s.preflightPath(chainID, blockNumber)
	data := &blockinputs.HeavyProverInputs{}
	reader, err := s.store.Load(ctx, path, &storeinputs.Headers{
		ContentType: storeinputs.ContentTypeJSON,
	})
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(reader).Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *ProverInputsStore) StoreProverInputs(ctx context.Context, data *blockinputs.ProverInputs, headers storeinputs.Headers) error {
	var buf bytes.Buffer
	switch s.format {
	case storeinputs.ContentTypeProtobuf:
		protoMsg := protoinputs.ToProto(data)
		protoBytes, err := proto.Marshal(protoMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf: %w", err)
		}
		buf.Write(protoBytes)
	case storeinputs.ContentTypeJSON:
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

	cfg := &Config{
		MultiConfig: multi.Config{},
	}

	storageType := headers.KeyValue["storage"]
	if storageType == "file" || storageType == "" {
		cfg.MultiConfig.FileConfig = &file.Config{DataDir: s.baseDir}
	}

	if storageType == "s3" || storageType == "" {
		keyPrefix := headers.KeyValue["key-prefix"] + data.ChainConfig.ChainID.String() + "/" + data.Block.Number.String()
		cfg.MultiConfig.S3Config = &s3.Config{
			Bucket:    headers.KeyValue["s3-bucket"],
			KeyPrefix: keyPrefix,
			ProviderConfig: &aws.ProviderConfig{
				Region: headers.KeyValue["region"],
				Credentials: &aws.CredentialsConfig{
					AccessKey: headers.KeyValue["access-key"],
					SecretKey: headers.KeyValue["secret-key"],
				},
			},
		}
	}

	compressStore, err := compress.New(compress.Config{
		ContentEncoding: headers.ContentEncoding,
		MultiConfig:     cfg.MultiConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to create compress store: %w", err)
	}

	path := s.proverPath(data.ChainConfig.ChainID.Uint64(), data.Block.Number.ToInt().Uint64(), headers)
	return compressStore.Store(ctx, path, bytes.NewReader(buf.Bytes()), &headers)
}

func (s *ProverInputsStore) LoadProverInputs(ctx context.Context, chainID, blockNumber uint64, headers storeinputs.Headers) (*blockinputs.ProverInputs, error) {
	// Initialize config based on storage type
	cfg := &Config{
		MultiConfig: multi.Config{},
	}

	storageType := headers.KeyValue["storage"]
	if storageType == "s3" {
		cfg.MultiConfig.S3Config = &s3.Config{
			Bucket: headers.KeyValue["s3-bucket"],
			ProviderConfig: &aws.ProviderConfig{
				Region: headers.KeyValue["region"],
				Credentials: &aws.CredentialsConfig{
					AccessKey: headers.KeyValue["access-key"],
					SecretKey: headers.KeyValue["secret-key"],
				},
			},
			KeyPrefix: headers.KeyValue["key-prefix"],
		}
	} else {
		cfg.MultiConfig.FileConfig = &file.Config{
			DataDir: s.baseDir,
		}
	}

	compressStore, err := compress.New(compress.Config{
		ContentEncoding: headers.ContentEncoding,
		MultiConfig:     cfg.MultiConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create compress store: %w", err)
	}

	path := s.proverPath(chainID, blockNumber, headers)
	reader, err := compressStore.Load(ctx, path, &headers)
	if err != nil {
		return nil, fmt.Errorf("failed to load data from store: %w", err)
	}

	data := &blockinputs.ProverInputs{}

	switch headers.ContentType {
	case storeinputs.ContentTypeJSON:
		if err := json.NewDecoder(reader).Decode(data); err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}
	case storeinputs.ContentTypeProtobuf:
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

func (s *HeavyProverInputsStore) preflightPath(chainID, blockNumber uint64) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%d", chainID), "preflight", fmt.Sprintf("%d.json", blockNumber))
}

func (s *ProverInputsStore) proverPath(chainID, blockNumber uint64, headers storeinputs.Headers) string {
	contentType, err := headers.GetContentType()
	if err != nil {
		return ""
	}

	keyPrefix := headers.KeyValue["key-prefix"]

	filename := fmt.Sprintf("%d.%s", blockNumber, contentType)
	if contentEncoding, err := headers.GetContentEncoding(); err == nil && contentEncoding != storeinputs.ContentEncodingPlain {
		filename = filename + "." + contentEncoding.String()
	}

	return filepath.Join(s.baseDir, keyPrefix, fmt.Sprintf("%d", chainID), filename)
}

// NewHeavyProverInputsStore creates a new HeavyProverInputsStore instance
func NewHeavyProverInputsStore(cfg *Config) (*HeavyProverInputsStore, error) {
	store := file.New(file.Config{
		DataDir: cfg.MultiConfig.FileConfig.DataDir,
	})

	return &HeavyProverInputsStore{
		store:   store,
		baseDir: cfg.MultiConfig.FileConfig.DataDir,
	}, nil
}
