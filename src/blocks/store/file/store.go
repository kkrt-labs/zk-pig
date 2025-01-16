package fileblockstore

// Implementation of BlockStore interface that stores the preflight and prover inputs in files.
//
// The preflight data is stored in at path `<base-dir>/<chainID>/preflight/<blockNumber>.json`
// The prover inputs are stored in a file named `<base-dir>/<chainID>/prover/<blockNumber>.json`

import (
	"compress/bzip2"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
	protoinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs/proto"
	filestore "github.com/kkrt-labs/kakarot-controller/src/blocks/store"
	"google.golang.org/protobuf/proto"
)

const (
	protobufFormat   = "protobuf"
	gzipCompression  = "gzip"
	flateCompression = "flate"
	bzip2Compression = "bzip2"
	zlibCompression  = "zlib"
)

type FileBlockStore struct {
	baseDir string
}

func New(baseDir string) *FileBlockStore {
	return &FileBlockStore{baseDir: baseDir}
}

func (s *FileBlockStore) StoreHeavyProverInputs(_ context.Context, inputs *blockinputs.HeavyProverInputs) error {
	path := s.preflightPath(inputs.ChainConfig.ChainID.Uint64(), inputs.Block.Number.ToInt().Uint64())
	return s.storeData(path, inputs, filestore.JSONFormat, filestore.NoCompression)
}

func (s *FileBlockStore) LoadHeavyProverInputs(_ context.Context, chainID, blockNumber uint64) (*blockinputs.HeavyProverInputs, error) {
	path := s.preflightPath(chainID, blockNumber)
	data := &blockinputs.HeavyProverInputs{}
	if err := s.loadData(path, data, filestore.JSONFormat, filestore.NoCompression); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *FileBlockStore) StoreProverInputs(_ context.Context, data *blockinputs.ProverInputs, format filestore.Format, compression filestore.Compression) error {
	path := s.proverPath(data.ChainConfig.ChainID.Uint64(), data.Block.Header.Number.ToInt().Uint64(), format, compression)
	return s.storeData(path, data, format, compression)
}

func (s *FileBlockStore) LoadProverInputs(_ context.Context, chainID, blockNumber uint64, format filestore.Format, compression filestore.Compression) (*blockinputs.ProverInputs, error) {
	path := s.proverPath(chainID, blockNumber, format, compression)
	data := &blockinputs.ProverInputs{}
	if err := s.loadData(path, data, format, compression); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *FileBlockStore) preflightPath(chainID, blockNumber uint64) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%d", chainID), "preflight", fmt.Sprintf("%d.json", blockNumber))
}

func (s *FileBlockStore) proverPath(chainID, blockNumber uint64, format filestore.Format, compression filestore.Compression) string {
	filename := fmt.Sprintf("%d.%s", blockNumber, format.String())
	if compression != filestore.NoCompression {
		filename = filename + "." + compression.String()
	}

	return filepath.Join(s.baseDir, fmt.Sprintf("%d", chainID), "prover-inputs", filename)
}

func getCompressWriter(w io.Writer, compression string) (io.WriteCloser, error) {
	switch compression {
	case gzipCompression:
		return gzip.NewWriterLevel(w, gzip.BestCompression)
	case flateCompression:
		return flate.NewWriter(w, flate.BestCompression)
	case zlibCompression:
		return zlib.NewWriterLevel(w, zlib.BestCompression)
	case bzip2Compression:
		return nil, fmt.Errorf("bzip2 compression not supported for writing")
	case "":
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported compression format: %s", compression)
	}
}

func getCompressReader(r io.Reader, compression string) (io.ReadCloser, error) {
	switch compression {
	case gzipCompression:
		return gzip.NewReader(r)
	case flateCompression:
		return flate.NewReader(r), nil
	case zlibCompression:
		return zlib.NewReader(r)
	case bzip2Compression:
		return io.NopCloser(bzip2.NewReader(r)), nil
	case "":
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported compression format: %s", compression)
	}
}

func (s *FileBlockStore) storeData(path string, data interface{}, format filestore.Format, compression filestore.Compression) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for file %s: %v", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", path, err)
	}
	defer file.Close()

	var writer io.Writer = file

	// Apply compression if specified
	compressor, err := getCompressWriter(file, compression.String())
	if err != nil {
		return err
	}
	if compressor != nil {
		writer = compressor
		defer func() {
			// First flush
			if c, ok := compressor.(interface{ Flush() error }); ok {
				if err := c.Flush(); err != nil {
					fmt.Printf("warning: failed to flush compressor: %v\n", err)
				}
			}
			// Then close
			if err := compressor.Close(); err != nil {
				fmt.Printf("warning: failed to close compressor: %v\n", err)
			}
		}()
	}

	// Write the data
	switch format {
	case filestore.ProtobufFormat:
		proverInputs, ok := data.(*blockinputs.ProverInputs)
		if !ok {
			return fmt.Errorf("data must be of type *blockinputs.ProverInputs for protobuf format")
		}
		protoMsg := protoinputs.ToProto(proverInputs)
		bytes, err := proto.Marshal(protoMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf data: %v", err)
		}
		if _, err := writer.Write(bytes); err != nil {
			return fmt.Errorf("failed to write protobuf data to file %s: %v", path, err)
		}
	case filestore.JSONFormat:
		if err := json.NewEncoder(writer).Encode(data); err != nil {
			return fmt.Errorf("failed to encode data to file %s: %v", path, err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format.String())
	}

	return nil
}

func (s *FileBlockStore) loadData(path string, data interface{}, format filestore.Format, compression filestore.Compression) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", path, err)
	}
	defer file.Close()

	var reader io.Reader = file

	// Apply compression if specified
	decompressor, err := getCompressReader(file, compression.String())
	if err != nil {
		return err
	}
	if decompressor != nil {
		reader = decompressor
		defer func() {
			if err := decompressor.Close(); err != nil {
				fmt.Printf("warning: failed to close decompressor: %v\n", err)
			}
		}()
	}

	switch format {
	case filestore.ProtobufFormat:
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
	case filestore.JSONFormat:
		if err := json.NewDecoder(reader).Decode(data); err != nil {
			return fmt.Errorf("failed to decode data from file %s: %v", path, err)
		}
	}
	return nil
}
