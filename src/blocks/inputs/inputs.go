package blockinputs

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc"
)

// ProverInputs is the block partial data expected by an EVM prover engine to execute & prove the block.
// It contains the minimal partial state & chain data necessary for the block execution
type ProverInputs struct {
	Block       *rpc.Block                      `json:"block"`       // Block data
	Ancestors   []*gethtypes.Header             `json:"ancestors"`   // Ancestors of the block that are accessed during the block execution
	ChainConfig *params.ChainConfig             `json:"chainConfig"` // Chain configuration
	Codes       []hexutil.Bytes                 `json:"codes"`       // Contract bytecodes used during the block execution
	PreState    []string                        `json:"preState"`    // Pre-state data, consisting in a list of MPT nodes
	AccessList  map[gethcommon.Address][]string `json:"accessList"`  // Access list of accounts and storage slots accessed during the block processing
}
