package blockstore

import (
	"context"

	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
)

type BlockStore interface {
	PreflightDataStore
	ProverInputsStore
}

type PreflightDataStore interface {
	// StorePreflightData stores the preflight data for a block.
	StorePreflightData(ctx context.Context, data *blockinputs.HeavyProverInputs) error

	// LoadPreflightData loads the preflight data for a block.
	LoadPreflightData(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.HeavyProverInputs, error)
}

type ProverInputsStore interface {
	// StoreProverInputs stores the prover inputs for a block.
	StoreProverInputs(ctx context.Context, data *blockinputs.ProverInputs) error

	// LoadProverInputs loads the prover inputs for a block.
	LoadProverInputs(ctx context.Context, chainID, blockNumber uint64) (*blockinputs.ProverInputs, error)
}
