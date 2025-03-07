package input

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// ProverInput contains the data expected by an EVM prover engine to execute & prove the block.
// It contains the minimal partial state & chain data necessary for processing the block and validating the final state.
type ProverInput struct {
	Version     string               `json:"version"`     // Prover Input version
	Blocks      []*Block             `json:"blocks"`      // Block to execute
	Witness     *Witness             `json:"witness"`     // Ancestors of the block that are accessed during the block execution
	ChainConfig *params.ChainConfig  `json:"chainConfig"` // Chain configuration
	AccessList  gethtypes.AccessList `json:"accessList"`  // Access list
}

type Witness struct {
	State     []hexutil.Bytes     `json:"state"`     // Partial pre-state, consisting in a list of MPT nodes
	Ancestors []*gethtypes.Header `json:"ancestors"` // Ancestors of the block that are accessed during the block execution
	Codes     []hexutil.Bytes     `json:"codes"`     // Contract bytecodes used during the block execution
}

type Block struct {
	Header       *gethtypes.Header        `json:"header"`
	Transactions []*gethtypes.Transaction `json:"transaction"`
	Uncles       []*gethtypes.Header      `json:"uncles"`
	Withdrawals  []*gethtypes.Withdrawal  `json:"withdrawals"`
}

func (b *Block) Block() *gethtypes.Block {
	return gethtypes.
		NewBlockWithHeader(b.Header).
		WithBody(
			gethtypes.Body{
				Transactions: b.Transactions,
				Uncles:       b.Uncles,
				Withdrawals:  b.Withdrawals,
			},
		)
}
