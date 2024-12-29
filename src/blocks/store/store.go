package blockstore

import (
	"context"

	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
)

type BlockStore interface {
	HeavyProverInputsStore
	ProverInputsStore
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
	LoadProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.ProverInputs, error)
}
