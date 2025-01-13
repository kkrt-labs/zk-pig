package blockinputs

import (
	"context"
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/evm"
	ethrpc "github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/state"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/trie"
	"github.com/kkrt-labs/kakarot-controller/pkg/log"
	"github.com/kkrt-labs/kakarot-controller/pkg/tag"
	"go.uber.org/zap"
)

// Preparer is the interface for preparing the prover inputs that serves as the input for the EVM prover engine.
// It runs a full "execution + final state validation" of the block ensuring that the necessary data is available.
// It bases on the heavy prover inputs generated at preflight to prepare the final prover inputs
type Preparer interface {
	// Prepare prepares the ProvableBlockInputs data for the EVM prover engine.
	Prepare(ctx context.Context, inputs *HeavyProverInputs) (*ProverInputs, error)
}

type preparer struct{}

// NewPreparer creates a new Preparer.
func NewPreparer() Preparer {
	return &preparer{}
}

// Prepare prepares the ProvableBlockInputs data for the EVM prover engine.
func (p *preparer) Prepare(ctx context.Context, data *HeavyProverInputs) (*ProverInputs, error) {
	ctx = tag.WithComponent(ctx, "prepare")
	ctx = tag.WithTags(
		ctx,
		tag.Key("chain.id").String(data.ChainConfig.ChainID.String()),
		tag.Key("block.number").Int64(data.Block.Number.ToInt().Int64()),
		tag.Key("block.hash").String(data.Block.Hash.Hex()),
	)

	inputs, err := p.prepare(ctx, data)
	if err != nil {
		log.LoggerFromContext(ctx).Error("Provable inputs preparation failed", zap.Error(err))
		return nil, err
	}
	log.LoggerFromContext(ctx).Info("Provable inputs preparation succeeded")

	return inputs, nil
}

type preparerContext struct {
	ctx      context.Context
	trackers *state.AccessTrackerManager
	stateDB  gethstate.Database
	hc       *core.HeaderChain
}

func (p *preparer) prepare(ctx context.Context, inputs *HeavyProverInputs) (*ProverInputs, error) {
	log.LoggerFromContext(ctx).Info("Process provable inputs preparation...")

	valCtx, err := p.prepareContext(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare validation context: %v", err)
	}

	if err := p.preparePreState(valCtx, inputs); err != nil {
		return nil, fmt.Errorf("failed to prefill validation database: %v", err)
	}

	execParams, err := p.prepareExecParams(valCtx, inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare validation exec params: %v", err)
	}

	if err := p.execute(valCtx, execParams); err != nil {
		return nil, fmt.Errorf("validation execution failed: %v", err)
	}

	return p.prepareProverInputs(valCtx, execParams), nil
}

func (p *preparer) prepareContext(ctx context.Context, inputs *HeavyProverInputs) (*preparerContext, error) {
	log.LoggerFromContext(ctx).Debug("Prepare context...")

	// --- Create necessary database and chain instances ---
	trackers := state.NewAccessTrackerManager()
	db := rawdb.NewMemoryDatabase()
	trieDB := triedb.NewDatabase(db, &triedb.Config{HashDB: &hashdb.Config{}})
	stateDB := state.NewAccessTrackerDatabase(state.NewModifiedTrieDatabase(trieDB, nil), trackers) // We use a modified trie database to track trie modifications

	hc, err := ethereum.NewChain(inputs.ChainConfig, stateDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create chain: %v", err)
	}

	return &preparerContext{
		ctx:      ctx,
		trackers: trackers,
		stateDB:  stateDB,
		hc:       hc,
	}, nil
}

func (p *preparer) preparePreState(ctx *preparerContext, inputs *HeavyProverInputs) error {
	log.LoggerFromContext(ctx.ctx).Info("Prepare pre-state...")

	// -- Preload the ancestors of the block into database ---
	ethereum.WriteHeaders(ctx.stateDB.TrieDB().Disk(), inputs.Ancestors...)

	// -- Preload the pre-state with the nodes obtained from the state proofs ---
	parentHeader := inputs.Ancestors[0]
	genesisHeader := ctx.hc.GetHeaderByNumber(0)

	nodeSet, err := trie.NodeSetFromStateTransitionProofs(parentHeader.Root, inputs.Block.Root, inputs.PreStateProofs, inputs.PostStateProofs)
	if err != nil {
		return fmt.Errorf("failed to create state nodes: %v", err)
	}

	err = ctx.stateDB.TrieDB().Update(parentHeader.Root, genesisHeader.Root, 0, nodeSet, triedb.NewStateSet())
	if err != nil {
		return fmt.Errorf("failed to update trie db with state nodes: %v", err)
	}

	// --- Preload the account bytecodes into the database ---
	codes := make([][]byte, 0)
	for _, code := range inputs.Codes {
		codes = append(codes, code)
	}
	ethereum.WriteCodes(ctx.stateDB.TrieDB().Disk(), codes...)

	return nil
}

func (p *preparer) prepareExecParams(ctx *preparerContext, inputs *HeavyProverInputs) (*evm.ExecParams, error) {
	log.LoggerFromContext(ctx.ctx).Debug("Prepare execution parameters...")

	parentHeader := inputs.Ancestors[0]
	preState, err := gethstate.New(parentHeader.Root, ctx.stateDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create pre-state from parent root %v: %v", parentHeader.Root, err)
	}

	return &evm.ExecParams{
		VMConfig: &vm.Config{
			StatelessSelfValidation: true,
		},
		Block:    inputs.Block.Block(),
		Validate: true, // We validate the block execution to ensure the result and final state are correct
		Chain:    ctx.hc,
		State:    preState,
	}, nil
}

func (p *preparer) execute(ctx *preparerContext, execParams *evm.ExecParams) error {
	log.LoggerFromContext(ctx.ctx).Info("Execute EVM...")
	_, err := evm.ExecutorWithTags("evm")(evm.ExecutorWithLog()(evm.NewExecutor())).Execute(ctx.ctx, execParams)
	if err != nil {
		return fmt.Errorf("failed to execute block: %v", err)
	}

	return nil
}

func (p *preparer) prepareProverInputs(ctx *preparerContext, execParams *evm.ExecParams) *ProverInputs {
	proverInputs := &ProverInputs{
		ChainConfig: execParams.Chain.Config(),
		Block:       new(ethrpc.Block).FromBlock(execParams.Block, execParams.Chain.Config()),
		Ancestors:   execParams.State.Witness().Headers,
	}

	witness := execParams.State.Witness()
	for code := range witness.Codes {
		proverInputs.Codes = append(proverInputs.Codes, []byte(code))
	}

	for node := range witness.State {
		blob := []byte(node)
		proverInputs.PreState = append(proverInputs.PreState, blob)
	}

	proverInputs.AccessList = make(map[gethcommon.Address][]hexutil.Bytes)
	tracker := ctx.trackers.GetAccessTracker(proverInputs.Ancestors[0].Root)
	for account := range tracker.Accounts {
		if storage, ok := tracker.Storage[account]; ok {
			proverInputs.AccessList[account] = []hexutil.Bytes{}
			for slot := range storage {
				proverInputs.AccessList[account] = append(proverInputs.AccessList[account], slot.Bytes())
			}
		}
	}

	return proverInputs
}
