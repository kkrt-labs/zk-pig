package generator

import (
	"context"
	"fmt"
	"math/big"

	gethtypes "github.com/ethereum/go-ethereum/core/types"
	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	"github.com/kkrt-labs/go-utils/tag"
	"github.com/kkrt-labs/zk-pig/src/steps"
	inputstore "github.com/kkrt-labs/zk-pig/src/store"
)

// Generator is a service that enables the generation of prover inpunts for EVM compatible blocks.
type Generator struct {
	ChainID *big.Int
	RPC     ethrpc.Client

	Preflighter steps.Preflight
	Preparer    steps.Preparer
	Executor    steps.Executor

	PreflightDataStore inputstore.PreflightDataStore
	ProverInputStore   inputstore.ProverInputStore
}

// Start starts the service.
func (s *Generator) Start(ctx context.Context) error {
	if s.RPC != nil {
		chainID, err := s.RPC.ChainID(ctx)
		if err != nil {
			return fmt.Errorf("failed to initialize RPC client: %v", err)
		}
		s.ChainID = chainID
	} else if s.ChainID == nil {
		return fmt.Errorf("no chain configuration provided")
	}

	return nil
}

func (s *Generator) Generate(ctx context.Context, blockNumber *big.Int) error {
	ctx = tag.WithComponent(ctx, "zkpig")
	ctx = tag.WithTags(ctx, tag.Key("block.number").Int64(blockNumber.Int64()))

	if s.RPC == nil {
		return fmt.Errorf("RPC not configured")
	}

	block, err := s.RPC.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch block: %v", err)
	}

	return s.generate(ctx, block)
}

func (s *Generator) generate(ctx context.Context, block *gethtypes.Block) error {
	data, err := s.preflight(ctx, block)
	if err != nil {
		return err
	}

	if err := s.prepare(ctx, data.Block.Number.ToInt()); err != nil {
		return err
	}

	if err := s.execute(ctx, data.Block.Number.ToInt()); err != nil {
		return err
	}

	return nil
}

// Preflight executes the preflight checks for the given block number.
// If requires the remote RPC to be configured and started
func (s *Generator) Preflight(ctx context.Context, blockNumber *big.Int) error {
	ctx = tag.WithComponent(ctx, "zkpig")

	if s.RPC == nil {
		return fmt.Errorf("RPC not configured")
	}

	block, err := s.RPC.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to fetch block: %v", err)
	}

	_, err = s.preflight(ctx, block)

	return err
}

func (s *Generator) preflight(ctx context.Context, block *gethtypes.Block) (*steps.PreflightData, error) {
	data, err := s.Preflighter.Preflight(ctx, block)
	if err != nil {
		return nil, fmt.Errorf("failed to execute preflight: %v", err)
	}

	if err = s.PreflightDataStore.StorePreflightData(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to store preflight data: %v", err)
	}

	return data, nil
}

func (s *Generator) Prepare(ctx context.Context, blockNumber *big.Int) error {
	ctx = tag.WithComponent(ctx, "zkpig")
	return s.prepare(ctx, blockNumber)
}

func (s *Generator) prepare(ctx context.Context, blockNumber *big.Int) error {
	if s.ChainID == nil {
		return fmt.Errorf("prepare: chain not configured")
	}

	data, err := s.PreflightDataStore.LoadPreflightData(ctx, s.ChainID.Uint64(), blockNumber.Uint64())
	if err != nil {
		return fmt.Errorf("prepare: failed to load preflight data: %v", err)
	}

	input, err := s.Preparer.Prepare(ctx, data)
	if err != nil {
		return fmt.Errorf("prepare: %v", err)
	}

	err = s.ProverInputStore.StoreProverInput(ctx, input)
	if err != nil {
		return fmt.Errorf("prepare: failed to store prover input: %v", err)
	}

	return nil
}

func (s *Generator) Execute(ctx context.Context, blockNumber *big.Int) error {
	ctx = tag.WithComponent(ctx, "zkpig")
	return s.execute(ctx, blockNumber)
}

func (s *Generator) execute(ctx context.Context, blockNumber *big.Int) error {
	if s.ChainID == nil {
		return fmt.Errorf("chain not configured")
	}

	input, err := s.ProverInputStore.LoadProverInput(ctx, s.ChainID.Uint64(), blockNumber.Uint64())
	if err != nil {
		return fmt.Errorf("failed to load provable inputs: %v", err)
	}

	_, err = s.Executor.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute block on provable inputs: %v", err)
	}

	return err
}

// Stop stops the service.
// Must be called to release resources.
func (s *Generator) Stop(ctx context.Context) error {
	return nil
}
