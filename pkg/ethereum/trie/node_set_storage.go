package trie

import (
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

// StorageProof holds proof for a storage slot
type StorageProof struct {
	Key   string      `json:"key"`
	Value hexutil.Big `json:"value,omitempty"`
	Proof []string    `json:"proof"`
}

// FromStorageRPC set StorageProof fields from a storage result returned by an RPC client and returns the StorageProof.
func StorageProofFromRPC(res *gethclient.StorageResult) *StorageProof {
	return &StorageProof{
		Key:   res.Key,
		Proof: res.Proof,
		Value: hexutil.Big(*res.Value),
	}
}

// StorageNodeSet is a wrapper around trienode.NodeSet that allows to add
// different types of nodes to a node set with proof verification
type StorageNodeSet struct {
	set *trienode.NodeSet
}

func NewStorageNodeSet(addr gethcommon.Address) *StorageNodeSet {
	return &StorageNodeSet{
		set: trienode.NewNodeSet(StorageTrieOwner(addr)),
	}
}

// Set returns the trienode.Set
func (ns *StorageNodeSet) Set() *trienode.NodeSet {
	return ns.set
}

// AddStorageNodes adds the storage nodes associated to the given storage proofs to the node set
// For each storage proof, it validates proof before adding the node to the set
func (ns *StorageNodeSet) AddStorageNodes(storageRoot gethcommon.Hash, storageProofs []*StorageProof) error {
	proofDB, keys, err := storageProofDBAndKeys(storageProofs)
	if err != nil {
		return err
	}

	return AddNodes(ns.set, storageRoot, proofDB, keys...)
}

func (ns *StorageNodeSet) AddStorageOrphanNodes(postRoot gethcommon.Hash, postProofs []*StorageProof) error {
	proofDB, keys, err := storageProofDBAndKeys(postProofs)
	if err != nil {
		return err
	}

	return AddOrphanNodes(ns.set, postRoot, proofDB, keys...)
}

func storageProofDBAndKeys(storageProofs []*StorageProof) (ethdb.KeyValueReader, [][]byte, error) {
	keys := make([][]byte, 0)
	proofDB := memorydb.New()
	for _, storageProof := range storageProofs {
		if len(storageProof.Proof) == 0 {
			continue
		}
		// Create the trie key for the storage slot
		key, err := hexutil.Decode(storageProof.Key)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode storage key %v: %w", storageProof.Key, err)
		}

		keys = append(keys, StorageTrieKey(key))

		// Populate the proof database with the storage proof
		err = trie.StoreHexProofs(storageProof.Proof, proofDB)
		if err != nil {
			return nil, nil, err
		}
	}

	return proofDB, keys, nil
}
