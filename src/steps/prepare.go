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
	"github.com/kkrt-labs/zk-pig/src/ethereum/trie"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=./mock/preparer.go -package=mocksteps github.com/kkrt-labs/zk-pig/src/steps Preparer

// Preparer is the interface for preparing the prover input that serves as the input for the EVM prover engine.
// It runs a full "execution + final state validation" of the block ensuring that the necessary data is available.
// It bases on the preflight data collected during preflight to prepare the final prover input
type Preparer interface {
	// Prepare prepares the ProvableBlockInputs data for the EVM prover engine.
	Prepare(ctx context.Context, data *PreflightData) (*input.ProverInput, error)
}

type preparer struct {
	evm evm.Executor
}

// NewPreparer creates a new Preparer.
func NewPreparer() Preparer {
	return NewPreparerFromEvm(
		evm.ExecutorWithTags("evm")(evm.ExecutorWithLog()(evm.NewExecutor())),
	)
}

// NewPreparerFromEvm creates a new Preparer from an EVM executor.
func NewPreparerFromEvm(e evm.Executor) Preparer {
	return &preparer{
		evm: e,
	}
}

// Prepare prepares the ProvableBlockInputs data for the EVM prover engine.
func (p *preparer) Prepare(ctx context.Context, data *PreflightData) (*input.ProverInput, error) {
	ctx = tag.WithComponent(ctx, "prepare")
	ctx = tag.WithTags(
		ctx,
		tag.Key("chain.id").String(data.ChainConfig.ChainID.String()),
		tag.Key("block.number").Int64(data.Block.Number.ToInt().Int64()),
		tag.Key("block.hash").String(data.Block.Hash.Hex()),
	)

	log.LoggerFromContext(ctx).Info("Start preparing prover input...")
	in, err := p.prepare(ctx, data)
	if err != nil {
		log.LoggerFromContext(ctx).Error("Prover input preparation failed", zap.Error(err))
		return nil, err
	}
	log.LoggerFromContext(ctx).Info("Prover input preparation succeeded")

	return in, nil
}

func (p *preparer) prepare(ctx context.Context, data *PreflightData) (*input.ProverInput, error) {
	stateDB, hc, err := p.prepareStateDBAndChain(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare state db and chain: %v", err)
	}

	// ___ Populate state database with state nodes ---
	parentHeader := hc.GetHeader(data.Block.Header.ParentHash, data.Block.Header.Number.ToInt().Uint64()-1)
	if parentHeader == nil {
		return nil, fmt.Errorf("missing parent header for block %q", data.Block.Header.Number.String())
	}

	preState, err := gethstate.New(parentHeader.Root, stateDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create pre-state from parent root %v: %v", parentHeader.Root, err)
	}

	execParams := &evm.ExecParams{
		VMConfig: &vm.Config{
			StatelessSelfValidation: true,
		},
		Block:    data.Block.Block(),
		Validate: true, // We validate the block execution to ensure the result and final state are correct
		Chain:    hc,
		State:    preState,
	}

	_, err = p.evm.Execute(ctx, execParams)
	if err != nil {
		return nil, fmt.Errorf("failed to execute block: %v", err)
	}

	return &input.ProverInput{
		ChainConfig: execParams.Chain.Config(),
		Blocks: []*input.Block{
			{
				Header:       execParams.Block.Header(),
				Transactions: execParams.Block.Transactions(),
				Uncles:       execParams.Block.Uncles(),
				Withdrawals:  execParams.Block.Withdrawals(),
			},
		},
		Witness: &input.Witness{
			Ancestors: execParams.State.Witness().Headers,
			Codes:     hexToHexBytes(execParams.State.Witness().Codes),
			State:     hexToHexBytes(execParams.State.Witness().State),
		},
	}, nil
}

func (p *preparer) prepareStateDBAndChain(_ context.Context, in *PreflightData) (gethstate.Database, *core.HeaderChain, error) {
	// --- Create in Memory database ---
	stateDB := gethstate.NewDatabase(
		triedb.NewDatabase(rawdb.NewMemoryDatabase(), &triedb.Config{HashDB: &hashdb.Config{}}),
		nil,
	) // We use a modified trie database to track trie modifications

	// -- Pre-populates database with Witness data ---
	ethereum.WriteHeaders(stateDB.TrieDB().Disk(), in.Ancestors...)
	ethereum.WriteCodes(stateDB.TrieDB().Disk(), hexBytesToBytes(in.Codes)...)

	// --- Create chain instance ---
	hc, err := ethereum.NewChain(in.ChainConfig, stateDB)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create chain: %v", err)
	}

	// ___ Populate state database with state nodes ---
	parentHeader := hc.GetHeader(in.Block.Header.ParentHash, in.Block.Header.Number.ToInt().Uint64()-1)
	if parentHeader == nil {
		return nil, nil, fmt.Errorf("missing parent header for block %q", in.Block.Header.Number.String())
	}

	genesisHeader := hc.GetHeaderByNumber(0)

	nodeSet, err := trie.NodeSetFromStateTransitionProofs(parentHeader.Root, in.Block.Root, in.PreStateProofs, in.PostStateProofs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create state nodes: %v", err)
	}

	err = stateDB.TrieDB().Update(parentHeader.Root, genesisHeader.Root, 0, nodeSet, triedb.NewStateSet())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update trie db with state nodes: %v", err)
	}

	return stateDB, hc, nil
}

func hexToHexBytes(hex map[string]struct{}) []hexutil.Bytes {
	bytes := make([]hexutil.Bytes, 0)
	for h := range hex {
		bytes = append(bytes, hexutil.Bytes(h))
	}
	return bytes
}
