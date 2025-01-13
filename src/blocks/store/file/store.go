package fileblockstore

// Implementation of BlockStore interface that stores the preflight and prover inputs in files.
//
// The preflight data is stored in at path `<base-dir>/<chainID>/preflight/<blockNumber>.json`
// The prover inputs are stored in a file named `<base-dir>/<chainID>/prover/<blockNumber>.json`

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
	protoinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs/proto"
	"google.golang.org/protobuf/proto"
)

const (
	protobufFormat = "protobuf"
	jsonFormat     = "json"
)

type FileBlockStore struct {
	baseDir string
}

func New(baseDir string) *FileBlockStore {
	return &FileBlockStore{baseDir: baseDir}
}

func (s *FileBlockStore) StoreHeavyProverInputs(_ context.Context, inputs *blockinputs.HeavyProverInputs) error {
	path := s.preflightPath(inputs.ChainConfig.ChainID.Uint64(), inputs.Block.Number.ToInt().Uint64())
	return s.storeData(path, inputs, jsonFormat)
}

func (s *FileBlockStore) LoadHeavyProverInputs(_ context.Context, chainID, blockNumber uint64) (*blockinputs.HeavyProverInputs, error) {
	path := s.preflightPath(chainID, blockNumber)
	data := &blockinputs.HeavyProverInputs{}
	if err := s.loadData(path, data, jsonFormat); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *FileBlockStore) StoreProverInputs(_ context.Context, data *blockinputs.ProverInputs, format string) error {
	path := s.proverPath(data.ChainConfig.ChainID.Uint64(), data.Block.Header.Number.ToInt().Uint64(), format)
	return s.storeData(path, data, format)
}

func (s *FileBlockStore) LoadProverInputs(_ context.Context, chainID, blockNumber uint64, format string) (*blockinputs.ProverInputs, error) {
	path := s.proverPath(chainID, blockNumber, format)
	data := &blockinputs.ProverInputs{}
	if err := s.loadData(path, data, format); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *FileBlockStore) preflightPath(chainID, blockNumber uint64) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%d", chainID), "preflight", fmt.Sprintf("%d.json", blockNumber))
}

func (s *FileBlockStore) proverPath(chainID, blockNumber uint64, format string) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%d", chainID), "prover-inputs", fmt.Sprintf("%d.%s", blockNumber, format))
}

func (s *FileBlockStore) storeData(path string, data interface{}, format string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for file %s: %v", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", path, err)
	}
	defer file.Close()

	switch format {
	case protobufFormat:
		proverInputs, ok := data.(*blockinputs.ProverInputs)
		if !ok {
			return fmt.Errorf("data must be of type *blockinputs.ProverInputs for protobuf format")
		}
		protoMsg := protoinputs.ToProto(proverInputs)
		bytes, err := proto.Marshal(protoMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf data: %v", err)
		}
		if _, err := file.Write(bytes); err != nil {
			return fmt.Errorf("failed to write protobuf data to file %s: %v", path, err)
		}
	case jsonFormat:
		if err := json.NewEncoder(file).Encode(data); err != nil {
			return fmt.Errorf("failed to encode data to file %s: %v", path, err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

func (s *FileBlockStore) loadData(path string, data interface{}, format string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", path, err)
	}
	defer file.Close()

	switch format {
	case protobufFormat:
		protoMsg := &protoinputs.ProverInputs{}
		bytes, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
		if err := proto.Unmarshal(bytes, protoMsg); err != nil {
			return fmt.Errorf("failed to unmarshal protobuf data: %v", err)
		}
		if proverInputs, ok := data.(*blockinputs.ProverInputs); ok {
			*proverInputs = *protoinputs.FromProto(protoMsg)
		} else {
			return fmt.Errorf("invalid data type: expected *blockinputs.ProverInputs")
		}
	case jsonFormat:
		if err := json.NewDecoder(file).Decode(data); err != nil {
			return fmt.Errorf("failed to decode data from file %s: %v", path, err)
		}
	}
	return nil
}
