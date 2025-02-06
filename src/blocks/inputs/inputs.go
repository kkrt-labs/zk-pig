package blockinputs

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/trie"
)

// ProverInputs contains the data expected by an EVM prover engine to execute & prove the block.
// It contains the minimal partial state & chain data necessary for processing the block and validating the final state.
type ProverInputs struct {
	Block       *ethrpc.Block                          `json:"block"`       // Block to execute
	Ancestors   []*gethtypes.Header                    `json:"ancestors"`   // Ancestors of the block that are accessed during the block execution
	ChainConfig *params.ChainConfig                    `json:"chainConfig"` // Chain configuration
	Codes       []hexutil.Bytes                        `json:"codes"`       // Contract bytecodes used during the block execution
	PreState    []hexutil.Bytes                        `json:"preState"`    // Partial pre-state, consisting in a list of MPT nodes
	AccessList  map[gethcommon.Address][]hexutil.Bytes `json:"accessList"`  // Access list of accounts and storage slots accessed during the block processing
}

// HeavyProverInputs contains data expected by an EVM prover engine to execute & prove the block.
// It contains the partial state & chain data necessary for processing the block and validating the final state.
// The format is convenient but sub-optimal as it contains duplicated data, it is an intermediate object necessary to generate the final ProverInputs.
type HeavyProverInputs struct {
	Block           *ethrpc.Block        `json:"block"`           // Block to execute
	Ancestors       []*gethtypes.Header  `json:"ancestors"`       // Ancestors of the block that are accessed during the block execution
	ChainConfig     *params.ChainConfig  `json:"chainConfig"`     // Chain configuration
	Codes           []hexutil.Bytes      `json:"codes"`           // Contract bytecodes used during the block execution
	PreStateProofs  []*trie.AccountProof `json:"preStateProofs"`  // Proofs of every accessed account and storage slot accessed during the block processing
	PostStateProofs []*trie.AccountProof `json:"postStateProofs"` // Proofs of every account and storage slot deleted during the block processing
}
