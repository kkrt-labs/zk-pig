package steps

import (
	"context"
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	"github.com/kkrt-labs/go-utils/log"
	"github.com/kkrt-labs/go-utils/tag"
	"github.com/kkrt-labs/zk-pig/src/ethereum"
	"github.com/kkrt-labs/zk-pig/src/ethereum/ethdb/rpcdb"
	"github.com/kkrt-labs/zk-pig/src/ethereum/evm"
	"github.com/kkrt-labs/zk-pig/src/ethereum/state"
	"github.com/kkrt-labs/zk-pig/src/ethereum/trie"
	"go.uber.org/zap"
)

// PreflightData contains data expected by an EVM prover engine to execute & prove the block.
// It contains the partial state & chain data necessary for processing the block and validating the final state.
// The format is convenient but sub-optimal as it contains duplicated data, it is an intermediate object necessary to generate the final ProverInput.
type PreflightData struct {
	Block           *ethrpc.Block        `json:"block"`           // Block to execute
	Ancestors       []*gethtypes.Header  `json:"ancestors"`       // Ancestors of the block that are accessed during the block execution
	ChainConfig     *params.ChainConfig  `json:"chainConfig"`     // Chain configuration
	Codes           []hexutil.Bytes      `json:"codes"`           // Contract bytecodes used during the block execution
	PreStateProofs  []*trie.AccountProof `json:"preStateProofs"`  // Proofs of every accessed account and storage slot accessed during the block processing
	PostStateProofs []*trie.AccountProof `json:"postStateProofs"` // Proofs of every account and storage slot deleted during the block processing
}

//go:generate mockgen -destination=./mock/preflight.go -package=mocksteps github.com/kkrt-labs/zk-pig/src/steps Preflight

// Preflight is the interface for the preflight block execution which consists of processing an EVM block without final state validation.
// It enables to collect necessary data for necessary for later full "block processing + final state validation".
// It outputs intermediary data that will be later used to prepare necessary pre-state data for the full block execution.
type Preflight interface {
	// Preflight executes a preflight block execution and returns the intermediate PreflightExecInputs data.
	Preflight(ctx context.Context, block *gethtypes.Block) (*PreflightData, error)
}

// preflight is the implementation of the Preflight interface using an RPC remote to fetch the state datas.
type preflight struct {
	remote ethrpc.Client

	chainCfg *params.ChainConfig

	evm evm.Executor
}

// NewPreflight creates a new RPC Preflight instance using the provided RPC client.
func NewPreflight(remote ethrpc.Client) Preflight {
	return NewPreflightFromEvm(
		evm.ExecutorWithTags("evm")(evm.ExecutorWithLog()(evm.NewExecutor())),
		remote,
	)
}

// NewPreflightFromEvm creates a new RPC Preflight instance using the provided EVM.
func NewPreflightFromEvm(e evm.Executor, remote ethrpc.Client) Preflight {
	return &preflight{
		remote: remote,
		evm:    e,
	}
}

// Init initializes preflight instance, ensuring the chain is supported.
func (pf *preflight) Start(ctx context.Context) error {
	return pf.init(ctx)
}

func (pf *preflight) init(ctx context.Context) error {
	var zero ethrpc.Client
	if pf.remote == nil || pf.remote == ethrpc.Client(nil) || pf.remote == zero {
		return fmt.Errorf("remote client not set")
	}

	// Fetch chain ID
	chainID, err := pf.remote.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch chain ID: %v", err)
	}

	pf.chainCfg, err = getChainConfig(chainID)
	if err != nil {
		return err
	}

	return nil
}

func (pf *preflight) configureDBAndChain(ctx context.Context) (*state.RPCDatabase, *core.HeaderChain, error) {
	db := rpcdb.HackWithContext(ctx, rawdb.NewMemoryDatabase(), pf.remote)
	trieDB := triedb.NewDatabase(db, &triedb.Config{HashDB: &hashdb.Config{}})
	rpcDB := state.HackWithContext(ctx, gethstate.NewDatabase(trieDB, nil), pf.remote)

	hc, err := ethereum.NewChain(pf.chainCfg, rpcDB)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create chain: %v", err)
	}

	return rpcDB, hc, nil
}

// Preflight executes a preflight block execution, that collect and returns the intermediary preflight data input.
func (pf *preflight) Preflight(ctx context.Context, block *gethtypes.Block) (*PreflightData, error) {
	// Addd preflight tags
	ctx = tag.WithComponent(ctx, "preflight")
	ctx = tag.WithTags(ctx, tag.Key("chain.id").String(pf.chainCfg.ChainID.String()))

	ctx = tag.WithTags(
		ctx,
		tag.Key("block.number").Int64(block.Number().Int64()),
		tag.Key("block.hash").String(block.Hash().Hex()),
	)

	// Execute preflight
	log.LoggerFromContext(ctx).Info("Run preflight to collect data from RPC node... (this may take a while)")
	data, err := pf.preflight(ctx, block)
	if err != nil {
		log.LoggerFromContext(ctx).Error("Preflight failed", zap.Error(err))
		return nil, err
	}

	log.LoggerFromContext(ctx).Info("Preflight succeeded")

	return data, nil
}

func (pf *preflight) preflight(ctx context.Context, block *gethtypes.Block) (*PreflightData, error) {
	db, hc, err := pf.configureDBAndChain(ctx)
	if err != nil {
		return nil, err
	}

	// Mark the parent block so the RPC database can work effectively
	parentHeader := hc.GetHeader(block.ParentHash(), block.Number().Uint64()-1)
	if parentHeader == nil {
		return nil, fmt.Errorf("preflight: missing parent header for block %q", block.Number().String())
	}
	db.MarkBlock(parentHeader)

	// Prepare state
	trackers := state.NewAccessTrackerManager()
	trackedDB := state.NewAccessTrackerDatabase(db, trackers)

	st, err := gethstate.New(parentHeader.Root, trackedDB)
	if err != nil {
		return nil, fmt.Errorf("preflight: failed to create state from parent root %v: %v", parentHeader.Root, err)
	}

	// EVM block execution
	execParams := &evm.ExecParams{
		Block:    block,
		Validate: false, // We do not validate on because we don't have a proper pre-state yet
		VMConfig: &vm.Config{
			StatelessSelfValidation: true, // We enable stateless self-validation so witness data is filled
		},
		Chain: hc,
		State: st,
	}
	_, err = pf.evm.Execute(ctx, execParams)
	if err != nil {
		return nil, fmt.Errorf("preflight: failed to execute block: %v", err)
	}

	// Prepare preflight data with ancestors and codes
	// Note: that at this stage the state witness does not contain enough nodes to derive the post-state root
	witness := execParams.State.Witness()
	data := &PreflightData{
		ChainConfig: pf.chainCfg,
		Block:       new(ethrpc.Block).FromBlock(block, pf.chainCfg),
		Ancestors:   witness.Headers,
	}
	for code := range witness.Codes {
		data.Codes = append(data.Codes, []byte(code))
	}

	// Fetch all necessary state proofs in order to derive the post-state root
	data.PreStateProofs, data.PostStateProofs, err = pf.fetchStateProofs(ctx, trackers, parentHeader, execParams)
	if err != nil {
		return nil, fmt.Errorf("preflight: failed to fetch state proofs: %v", err)
	}

	return data, nil
}

// fetchStateProofs for all accounts and storage slots that were accessed during the block execution
// It fetches the state proofs both at the initial state (parent state) and at the final state
func (pf *preflight) fetchStateProofs(ctx context.Context, trackers *state.AccessTrackerManager, parentHeader *gethtypes.Header, execParams *evm.ExecParams) (preStateProofs, postStateProofs []*trie.AccountProof, err error) {
	finalState := execParams.State
	tracker := trackers.GetAccessTracker(parentHeader.Root)
	for addr, accountAccessTracker := range tracker.Accounts {
		var (
			slots       = []string{}
			deletedSlot = []string{}
		)

		for slot, preStateValue := range accountAccessTracker.Storage {
			slots = append(slots, slot.Hex())
			if (preStateValue != gethcommon.Hash{}) && (finalState.GetState(addr, slot) == gethcommon.Hash{}) {
				deletedSlot = append(deletedSlot, slot.Hex())
			}
		}

		// Get proofs for every accounts on the initial state (parent state)
		acc, err := pf.remote.GetProof(ctx, addr, slots, parentHeader.Number)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get proof for account %v: %v", addr, err)
		}
		preStateProofs = append(preStateProofs, trie.AccountProofFromRPC(acc))

		// Also get necessary proofs at final state
		if len(deletedSlot) == 0 && !finalState.HasSelfDestructed(addr) {
			// Account was not deleted so we don't need to fetch post-state proofs for it
			continue
		}

		// Also get proofs at final state for deleted accounts & slots
		acc, err = pf.remote.GetProof(ctx, addr, deletedSlot, execParams.Block.Number())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get proof for account %v: %v", addr, err)
		}
		postStateProofs = append(postStateProofs, trie.AccountProofFromRPC(acc))
	}

	return preStateProofs, postStateProofs, nil
}

func (pf *preflight) Stop(_ context.Context) error {
	return nil
}
