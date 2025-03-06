package steps

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/kkrt-labs/go-utils/log"
	"github.com/kkrt-labs/go-utils/tag"
	"github.com/kkrt-labs/zk-pig/src/ethereum"
	"github.com/kkrt-labs/zk-pig/src/ethereum/evm"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=./mock/executor.go -package=mocksteps github.com/kkrt-labs/zk-pig/src/steps Executor

// Executor is the interface for EVM execution on provable inputs.
// It runs a full "execution + final state validation" of the block
// It is primarily meant to validate that the provable inputs are correct and enable proper EVM execution.
type Executor interface {
	// Execute runs a full EVM block execution on provable inputs
	Execute(ctx context.Context, input *input.ProverInput) (*core.ProcessResult, error)
}

type executor struct {
	evm evm.Executor
}

// NewExecutor creates a new instance of the Executor.
func NewExecutor() Executor {
	return NewExecutorFromEvm(evm.ExecutorWithTags("evm")(evm.ExecutorWithLog()(evm.NewExecutor())))
}

// NewExecutorFromEvm creates a new instance of the BaseExecutor from an existing EVM.
func NewExecutorFromEvm(e evm.Executor) Executor {
	return &executor{
		evm: e,
	}
}

// Execute runs the ProvableBlockInputs data for the EVM prover engine.
func (e *executor) Execute(ctx context.Context, in *input.ProverInput) (*core.ProcessResult, error) {
	if len(in.Blocks) == 0 {
		return nil, fmt.Errorf("no blocks provided")
	}

	block := in.Blocks[0]

	ctx = tag.WithComponent(ctx, "execute")
	ctx = tag.WithTags(
		ctx,
		tag.Key("chain.id").String(in.ChainConfig.ChainID.String()),
		tag.Key("block.number").Int64(block.Header.Number.Int64()),
		tag.Key("block.hash").String(block.Header.Hash().Hex()),
	)

	log.LoggerFromContext(ctx).Info("Execute block and validate state transition by basing on prover input...")
	res, err := e.execute(ctx, in)
	if err != nil {
		log.LoggerFromContext(ctx).Error("Block execution failed", zap.Error(err))
		return res, err
	}

	log.LoggerFromContext(ctx).Info("Block execution succeeded")

	return res, err
}

func (e *executor) execute(ctx context.Context, in *input.ProverInput) (*core.ProcessResult, error) {
	stateDB, hc, err := e.prepareStateDBAndChain(in)
	if err != nil {
		return nil, fmt.Errorf("execute: failed to prepare state db and chain: %v", err)
	}

	parentHeader := hc.GetHeader(in.Blocks[0].Header.ParentHash, in.Blocks[0].Header.Number.Uint64()-1)
	if parentHeader == nil {
		return nil, fmt.Errorf("execute: missing parent header for block %q", in.Blocks[0].Header.Number.String())
	}

	preState, err := gethstate.New(parentHeader.Root, stateDB)
	if err != nil {
		return nil, fmt.Errorf("execute: failed to create pre-state from parent root %v: %v", parentHeader.Root, err)
	}

	execParams := &evm.ExecParams{
		VMConfig: &vm.Config{
			StatelessSelfValidation: true,
		},
		Block:    in.Blocks[0].Block(),
		Validate: true, // We validate the block execution to ensure the result and final state are correct
		Chain:    hc,
		State:    preState,
	}

	res, err := e.evm.Execute(ctx, execParams)
	if err != nil {
		return res, fmt.Errorf("execute: %v", err)
	}

	return res, nil
}

func (e *executor) prepareStateDBAndChain(in *input.ProverInput) (gethstate.Database, *core.HeaderChain, error) {
	// --- Create in Memory database ---
	stateDB := gethstate.NewDatabase(
		triedb.NewDatabase(rawdb.NewMemoryDatabase(), &triedb.Config{HashDB: &hashdb.Config{}}),
		nil,
	) // We use a modified trie database to track trie modifications

	// -- Pre-populates database with Witness data ---
	ethereum.WriteHeaders(stateDB.TrieDB().Disk(), in.Witness.Ancestors...)
	ethereum.WriteCodes(stateDB.TrieDB().Disk(), hexBytesToBytes(in.Witness.Codes)...)
	ethereum.WriteNodesToHashDB(stateDB.TrieDB().Disk(), hexBytesToBytes(in.Witness.State)...)

	hc, err := ethereum.NewChain(in.ChainConfig, stateDB)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create chain: %v", err)
	}

	return stateDB, hc, nil
}

func hexBytesToBytes(hex []hexutil.Bytes) [][]byte {
	bytes := make([][]byte, 0)
	for _, b := range hex {
		bytes = append(bytes, b)
	}
	return bytes
}
