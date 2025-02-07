package trie

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

// AccountProof holds proofs for an Ethereum account and optionnaly its storage
type AccountProof struct {
	Address     gethcommon.Address `json:"address"`
	Proof       []string           `json:"accountProof"`
	Balance     hexutil.Big        `json:"balance,omitempty"`
	CodeHash    gethcommon.Hash    `json:"codeHash,omitempty"`
	Nonce       uint64             `json:"nonce,omitempty"`
	StorageHash gethcommon.Hash    `json:"storageHash,omitempty"`
	Storage     []*StorageProof    `json:"storageProof,omitempty"`
}

// AccountProofFromRPC sets AccountProof fields from an account result returned by an RPC client and returns the AccountProof.
func AccountProofFromRPC(res *gethclient.AccountResult) *AccountProof {
	proof := &AccountProof{
		Address:     res.Address,
		Proof:       res.AccountProof,
		Balance:     hexutil.Big(*res.Balance),
		CodeHash:    res.CodeHash,
		Nonce:       res.Nonce,
		StorageHash: res.StorageHash,
	}

	for _, slot := range res.StorageProof {
		proof.Storage = append(proof.Storage, StorageProofFromRPC(&slot))
	}

	return proof
}

// AccountsNodeSet is a wrapper around trienode.NodeSet that allows to add
// accounts to a MPT node set with proof verification
type AccountsNodeSet struct {
	set *trienode.NodeSet
}

// NewProvedNodeSet creates a new ProvedNodeSet
// - owner of the nodeset: empty for the account trie and the owning account address hash for storage tries.
func NewAccountNodeSet() *AccountsNodeSet {
	return &AccountsNodeSet{
		set: trienode.NewNodeSet(gethcommon.Hash{}),
	}
}

// Set returns the trienode.Set
func (ns *AccountsNodeSet) Set() *trienode.NodeSet {
	return ns.set
}

// AddAccountNodes adds the account nodes associated to the given account proofs to the node set
// For each account proof, it validates proof before adding the node to the set
func (ns *AccountsNodeSet) AddAccountNodes(accountRoot gethcommon.Hash, accountProofs []*AccountProof) error {
	proofDB, keys, err := accountsProofDBAndKeys(accountProofs)
	if err != nil {
		return nil
	}

	return AddNodes(ns.set, accountRoot, proofDB, keys...)
}

func (ns *AccountsNodeSet) AddAccountOrphanNodes(accountRoot gethcommon.Hash, accountProofs []*AccountProof) error {
	proofDB, keys, err := accountsProofDBAndKeys(accountProofs)
	if err != nil {
		return nil
	}

	return AddOrphanNodes(ns.set, accountRoot, proofDB, keys...)
}

func accountsProofDBAndKeys(accountProofs []*AccountProof) (ethdb.KeyValueReader, [][]byte, error) {
	keys := make([][]byte, 0)
	proofDB := memorydb.New()
	for _, accountProof := range accountProofs {
		// Create the trie key for the account
		keys = append(keys, AccountTrieKey(accountProof.Address))

		// Populate the proof database with the account proof
		err := trie.StoreHexProofs(accountProof.Proof, proofDB)
		if err != nil {
			return nil, nil, err
		}
	}

	return proofDB, keys, nil
}
