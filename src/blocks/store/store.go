package blockstore

import (
	"context"

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

type HeavyProverInputsStore interface {
	// StoreHeavyProverInputs stores the heavy prover inputs for a block.
	StoreHeavyProverInputs(ctx context.Context, inputs *blockinputs.HeavyProverInputs) error

	// LoadHeavyProverInputs loads tthe heavy prover inputs for a block.
	LoadHeavyProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.HeavyProverInputs, error)
}

type ProverInputsStore interface {
	// StoreProverInputs stores the prover inputs for a block.
	// format can be "protobuf" or "json"
	StoreProverInputs(ctx context.Context, inputs *blockinputs.ProverInputs, format Format) error

	// LoadProverInputs loads the prover inputs for a block.
	// format can be "protobuf" or "json"
	LoadProverInputs(ctx context.Context, chainID, blockNumber uint64, format Format) (*blockinputs.ProverInputs, error)
}

func (f Format) String() string {
	switch f {
	case JSONFormat:
		return "json"
	case ProtobufFormat:
		return "protobuf"
	default:
		return "json"
	}
}
