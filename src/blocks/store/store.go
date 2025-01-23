package blockstore

import (
	"context"

	filestore "github.com/kkrt-labs/kakarot-controller/pkg/store"
	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
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
	StoreProverInputs(ctx context.Context, inputs *blockinputs.ProverInputs, headers filestore.Headers) error

	// LoadProverInputs loads the prover inputs for a block.
	// format can be "protobuf" or "json"
	LoadProverInputs(ctx context.Context, chainID, blockNumber uint64, headers filestore.Headers) (*blockinputs.ProverInputs, error)
}
