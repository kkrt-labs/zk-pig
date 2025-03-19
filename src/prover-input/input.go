package input

import (
	"encoding/json"
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// ProverInput contains the data expected by an EVM prover engine to execute & prove the block.
// It contains the minimal partial state & chain data necessary for processing the block and validating the final state.
type ProverInput struct {
	Version     string              `json:"version"`         // Prover Input version
	Blocks      []*Block            `json:"blocks"`          // Block to execute
	Witness     *Witness            `json:"witness"`         // Ancestors of the block that are accessed during the block execution
	ChainConfig *params.ChainConfig `json:"chainConfig"`     // Chain configuration
	Extra       *Extra              `json:"extra,omitempty"` // Extra data
}

type Witness struct {
	State     [][]byte            // Partial pre-state, consisting in a list of MPT nodes
	Ancestors []*gethtypes.Header // Ancestors of the block that are accessed during the block execution
	Codes     [][]byte            // Contract bytecodes used during the block execution
}

type witnessMarshaling struct {
	State     []hexutil.Bytes     `json:"state"`
	Ancestors []*gethtypes.Header `json:"ancestors"`
	Codes     []hexutil.Bytes     `json:"codes"`
}

func (w Witness) MarshalJSON() ([]byte, error) {
	return json.Marshal(witnessMarshaling{
		State:     bytesToHex(w.State),
		Ancestors: w.Ancestors,
		Codes:     bytesToHex(w.Codes),
	})
}

func bytesToHex(b [][]byte) []hexutil.Bytes {
	hexBytes := make([]hexutil.Bytes, len(b))
	for i, b := range b {
		hexBytes[i] = hexutil.Bytes(b)
	}
	return hexBytes
}

func (w *Witness) UnmarshalJSON(b []byte) error {
	var m witnessMarshaling
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	w.State = hexToBytes(m.State)
	w.Ancestors = m.Ancestors
	w.Codes = hexToBytes(m.Codes)

	return nil
}

func hexToBytes(h []hexutil.Bytes) [][]byte {
	bytes := make([][]byte, len(h))
	for i, b := range h {
		bytes[i] = []byte(b)
	}
	return bytes
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

// Extra contains additional data that can be included in the Prover Input.
type Extra struct {
	AccessList gethtypes.AccessList                 `json:"accessList,omitempty"` // Access list of addresses and storage slots that were accessed during block execution
	Committed  [][]byte                             `json:"committed,omitempty"`  // Nodes committed during block execution
	StateDiffs []*StateDiff                         `json:"stateDiffs,omitempty"` // State diffs for accounts that have changes during block execution
	PreState   map[gethcommon.Address]*AccountState `json:"preState,omitempty"`   // Pre-state for accounts that have changes during block execution
}

// StateDiff represents a difference in the state of an account.
type StateDiff struct {
	Address     gethcommon.Address `json:"address"`               // Address of the account that has changed
	PreAccount  *Account           `json:"preAccount,omitempty"`  // Pre-state account, nil if the account if the account was created
	PostAccount *Account           `json:"postAccount,omitempty"` // Post-state account, nil if the account was desctructed
	Storage     []*StorageDiff     `json:"storage,omitempty"`     // Storage diffs for the account
}

// Account represents an account in the state.
type Account struct {
	Balance     *big.Int        `json:"balance"`
	CodeHash    gethcommon.Hash `json:"codeHash"`
	Nonce       uint64          `json:"nonce"`
	StorageHash gethcommon.Hash `json:"storageHash"`
}

// StorageDiff represents a difference in the storage of an account.
type StorageDiff struct {
	Slot      gethcommon.Hash `json:"storageKey"`
	PreValue  gethcommon.Hash `json:"preValue,omitempty"`
	PostValue gethcommon.Hash `json:"postValue,omitempty"`
}

// AccountState represents the state of an account.
type AccountState struct {
	Balance     *big.Int
	CodeHash    gethcommon.Hash
	Code        []byte
	Nonce       uint64
	StorageHash gethcommon.Hash
	Storage     map[gethcommon.Hash]gethcommon.Hash
}

type accountStateMarshaling struct {
	Balance     *hexutil.Big                        `json:"balance"`
	CodeHash    gethcommon.Hash                     `json:"codeHash"`
	Code        hexutil.Bytes                       `json:"code,omitempty"`
	Nonce       uint64                              `json:"nonce"`
	StorageHash gethcommon.Hash                     `json:"storageHash"`
	Storage     map[gethcommon.Hash]gethcommon.Hash `json:"storage,omitempty"`
}

func (a *AccountState) MarshalJSON() ([]byte, error) {
	return json.Marshal(accountStateMarshaling{
		Balance:     (*hexutil.Big)(a.Balance),
		CodeHash:    a.CodeHash,
		Code:        a.Code,
		Nonce:       a.Nonce,
		StorageHash: a.StorageHash,
		Storage:     a.Storage,
	})
}

func (a *AccountState) UnmarshalJSON(b []byte) error {
	var m accountStateMarshaling
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	a.Balance = (*big.Int)(m.Balance)
	a.CodeHash = m.CodeHash
	a.Code = m.Code
	a.Nonce = m.Nonce
	a.StorageHash = m.StorageHash
	a.Storage = m.Storage

	return nil
}
