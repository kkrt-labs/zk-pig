package blockstore

import (
	"context"
	"fmt"

	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
)

type BlockStore interface {
	HeavyProverInputsStore
	ProverInputsStore
}

type Format int

const (
	JSONFormat Format = iota
	ProtobufFormat
)

type Compression int

const (
	GzipCompression Compression = iota
	FlateCompression
	ZlibCompression
	NoCompression
)

func (f Format) String() string {
	switch f {
	case JSONFormat:
		return "json"
	case ProtobufFormat:
		return "protobuf"
	default:
		return ""
	}
}

func (c Compression) String() string {
	switch c {
	case GzipCompression:
		return "gzip"
	case FlateCompression:
		return "flate"
	case ZlibCompression:
		return "zlib"
	case NoCompression:
		return ""
	}
	return ""
}

var formats = map[string]Format{
	"json":     JSONFormat,
	"protobuf": ProtobufFormat,
}

var compressions = map[string]Compression{
	"gzip":  GzipCompression,
	"flate": FlateCompression,
	"zlib":  ZlibCompression,
	"none":  NoCompression,
	"":      NoCompression,
}

func ParseFormat(formatStr string) (Format, error) {
	if f, ok := formats[formatStr]; ok {
		return f, nil
	}
	return 0, fmt.Errorf("unsupported store format %q", formatStr)
}

func ParseCompression(compressionStr string) (Compression, error) {
	if c, ok := compressions[compressionStr]; ok {
		return c, nil
	}
	return 0, fmt.Errorf("unsupported store compression %q", compressionStr)
}

type HeavyProverInputsStore interface {
	// StoreHeavyProverInputs stores the heavy prover inputs for a block.
	StoreHeavyProverInputs(ctx context.Context, inputs *blockinputs.HeavyProverInputs) error

	// LoadHeavyProverInputs loads tthe heavy prover inputs for a block.
	LoadHeavyProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.HeavyProverInputs, error)
}

type ProverInputsStore interface {
	// StoreProverInputs stores the prover inputs for a block.
	StoreProverInputs(ctx context.Context, inputs *blockinputs.ProverInputs, format Format, compression Compression) error

	// LoadProverInputs loads the prover inputs for a block.
	// format can be "protobuf" or "json"
	LoadProverInputs(ctx context.Context, chainID, blockNumber uint64, format Format, compression Compression) (*blockinputs.ProverInputs, error)
}
