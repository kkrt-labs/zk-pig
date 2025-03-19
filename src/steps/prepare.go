package steps

import (
	"bytes"
	"context"
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/kkrt-labs/go-utils/log"
	"github.com/kkrt-labs/go-utils/tag"
	"github.com/kkrt-labs/zk-pig/src/ethereum"
	"github.com/kkrt-labs/zk-pig/src/ethereum/evm"
	"github.com/kkrt-labs/zk-pig/src/ethereum/state"
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

	includeOpt Include
}

type PrepareOption func(*preparer) error

// NewPreparer creates a new Preparer.
func NewPreparer(opts ...PrepareOption) (Preparer, error) {
	return NewPreparerFromEvm(
		evm.ExecutorWithTags("evm")(evm.ExecutorWithLog()(evm.NewExecutor())),
		opts...,
	)
}

// NewPreparerFromEvm creates a new Preparer from an EVM executor.
func NewPreparerFromEvm(e evm.Executor, opts ...PrepareOption) (Preparer, error) {
	p := &preparer{
		evm: e,
	}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	return p, nil
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

	trackers := state.NewAccessTrackerManager()
	trackedDB := state.NewAccessTrackerDatabase(stateDB, trackers)

	// ___ Populate state database with state nodes ---
	parentHeader := hc.GetHeader(data.Block.Header.ParentHash, data.Block.Header.Number.ToInt().Uint64()-1)
	if parentHeader == nil {
		return nil, fmt.Errorf("missing parent header for block %q", data.Block.Header.Number.String())
	}

	preState, err := gethstate.New(parentHeader.Root, trackedDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create pre-state from parent root %v: %v", parentHeader.Root, err)
	}

	execParams := &evm.ExecParams{
		VMConfig: &vm.Config{
			StatelessSelfValidation: true,
		},
		Block:    data.Block.Block(),
		Validate: true, // We validate the block execution to ensure the result and final state are correct
		Commit:   p.include(IncludeCommitted),
		Chain:    hc,
		State:    preState,
	}

	_, err = p.evm.Execute(ctx, execParams)
	if err != nil {
		return nil, fmt.Errorf("failed to execute block: %v", err)
	}

	extra := new(input.Extra)

	if p.include(IncludeAccessList) {
		tracker := trackers.GetAccessTracker(parentHeader.Root)
		for addr, accountAccessTracker := range tracker.Accounts {
			accessTuple := gethtypes.AccessTuple{
				Address:     addr,
				StorageKeys: make([]gethcommon.Hash, 0),
			}

			for slot := range accountAccessTracker.Storage {
				accessTuple.StorageKeys = append(accessTuple.StorageKeys, slot)
			}
			extra.AccessList = append(extra.AccessList, accessTuple)
		}
	}

	if p.include(IncludeStateDiffs) {
		tracker := trackers.GetAccessTracker(parentHeader.Root)
		for addr, accountAccessTracker := range tracker.Accounts {
			preAcc := accountAccessTracker.Account
			postAcc := getAccount(execParams.State, addr)

			if !accountHasChanged(preAcc, postAcc) {
				continue
			}

			stateDiff := &input.StateDiff{
				Address:     addr,
				PreAccount:  toAccount(preAcc),
				PostAccount: toAccount(postAcc),
			}

			if accountRootHasChanged(preAcc, postAcc) {
				for slot, preValue := range accountAccessTracker.Storage {
					postValue := execParams.State.GetState(addr, slot)
					if preValue != postValue {
						stateDiff.Storage = append(stateDiff.Storage, &input.StorageDiff{
							Slot:      slot,
							PreValue:  preValue,
							PostValue: postValue,
						})
					}
				}
			}

			extra.StateDiffs = append(extra.StateDiffs, stateDiff)
		}
	}

	if p.include(IncludeCommitted) {
		extra.Committed = witnessToBytes(execParams.State.Witness().Committed)
	}

	if p.include(IncludePreState) {
		extra.PreState = make(map[gethcommon.Address]*input.AccountState)
		tracker := trackers.GetAccessTracker(parentHeader.Root)
		for addr, accountAccessTracker := range tracker.Accounts {
			if accountAccessTracker.Account == nil {
				extra.PreState[addr] = nil
				continue
			}

			extra.PreState[addr] = &input.AccountState{
				Balance:     accountAccessTracker.Account.Balance.ToBig(),
				CodeHash:    gethcommon.BytesToHash(accountAccessTracker.Account.CodeHash),
				Code:        execParams.State.GetCode(addr),
				Nonce:       accountAccessTracker.Account.Nonce,
				StorageHash: accountAccessTracker.Account.Root,
				Storage:     accountAccessTracker.Storage,
			}
		}
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
			Codes:     witnessToBytes(execParams.State.Witness().Codes),
			State:     witnessToBytes(execParams.State.Witness().State),
		},
		Extra: extra,
	}, nil
}

func (p *preparer) include(opt Include) bool {
	return p.includeOpt.Include(opt)
}

func getAccount(state *gethstate.StateDB, addr gethcommon.Address) *gethtypes.StateAccount {
	balance := state.GetBalance(addr)
	nonce := state.GetNonce(addr)
	codeHash := state.GetCodeHash(addr)
	root := state.GetStorageRoot(addr)
	if balance.IsZero() && nonce == 0 && codeHash == (gethcommon.Hash{}) && root == (gethcommon.Hash{}) {
		return nil
	}

	return &gethtypes.StateAccount{
		Balance:  balance,
		Nonce:    nonce,
		CodeHash: codeHash.Bytes(),
		Root:     root,
	}
}

func accountRootHasChanged(preAcc, postAcc *gethtypes.StateAccount) bool {
	if preAcc == nil && postAcc == nil {
		return false
	}

	if preAcc != nil && postAcc == nil || preAcc == nil && postAcc != nil {
		return true
	}

	return preAcc.Root != postAcc.Root
}

func accountHasChanged(preAcc, postAcc *gethtypes.StateAccount) bool {
	if preAcc == nil && postAcc == nil {
		return false
	}

	if preAcc != nil && postAcc == nil || preAcc == nil && postAcc != nil {
		return true
	}

	return preAcc.Balance.Cmp(postAcc.Balance) != 0 ||
		preAcc.Nonce != postAcc.Nonce ||
		!bytes.Equal(preAcc.CodeHash, postAcc.CodeHash) ||
		preAcc.Root != postAcc.Root
}

func toAccount(acc *gethtypes.StateAccount) *input.Account {
	if acc == nil {
		return nil
	}

	return &input.Account{
		Balance:     acc.Balance.ToBig(),
		CodeHash:    gethcommon.BytesToHash(acc.CodeHash),
		Nonce:       acc.Nonce,
		StorageHash: acc.Root,
	}
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

func witnessToBytes(hex map[string]struct{}) [][]byte {
	bytes := make([][]byte, 0)
	for h := range hex {
		bytes = append(bytes, []byte(h))
	}
	return bytes
}
