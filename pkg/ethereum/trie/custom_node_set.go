package trie

import (
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
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

// ProvedNodeSet is a wrapper around trienode.NodeSet that allows to add
// different types of nodes to a node set with proof verification
type ProvedNodeSet struct {
	set *trienode.NodeSet
}

// NewProvedNodeSet creates a new ProvedNodeSet
// - owner of the nodeset: empty for the account trie and the owning account address hash for storage tries.
func NewProvedNodeSet(owner gethcommon.Hash) *ProvedNodeSet {
	return NewProvedNodeSetFromSet(trienode.NewNodeSet(owner))
}

// NewProvedNodeSet creates a new ProvedNodeSet
// - owner of the nodeset: empty for the account trie and the owning account address hash for storage tries.
func NewProvedAccountNodeSet() *ProvedNodeSet {
	return NewProvedNodeSet(gethcommon.Hash{})
}

// NewProvedStorageNodeSet creates a new ProvedNodeSet
func NewProvedStorageNodeSet(addr gethcommon.Address) *ProvedNodeSet {
	return NewProvedNodeSet(StorageTrieOwner(addr))
}

// NewProvedNodeSetFromSet creates a new ProvedNodeSet from an existing trienode.NodeSet
func NewProvedNodeSetFromSet(set *trienode.NodeSet) *ProvedNodeSet {
	return &ProvedNodeSet{set: set}
}

// Set returns the trienode.Set
func (ns *ProvedNodeSet) Set() *trienode.NodeSet {
	return ns.set
}

// AddNodes adds the nodes associated to the given keys to the node set
// For each key, it validates proof before adding the node to the set
// - root is the state root hash
// - proofDB is a database containing all the necessary trie nodes (keyed by node hash)
// - keys is a list of keys to insert into the node set
func (ns *ProvedNodeSet) AddNodes(root gethcommon.Hash, proofDB ethdb.KeyValueReader, keys ...[]byte) error {
	for _, key := range keys {
		_, _, _, err := VerifyProofWithReporting(
			root,
			key,
			proofDB,
			func(hash gethcommon.Hash, key, node []byte) {
				ns.set.AddNode(key, trienode.New(hash, node))
			},
			ns.set.AddLeaf,
		)
		if err != nil {
			return fmt.Errorf("failed to add proof for key %x: %v", key, err)
		}
	}

	return nil
}

// AddOrphanNodes adds orphan nodes to the node set for keys deleted during a state transition, ensuring
// the node set contains all necessary nodes for proper key deletion.
//
// This is required because deleting a key may reduce the trie, and orphan nodes are needed to handle
// the reduction when removing a branch node.
//
// Assumptions:
// - NodeSet holds nodes of the pre-state trie.
// - postRoot is the root hash of the post-state trie.
// - postProofDB is a database containing trie nodes for the post-state.
// - keys are keys deleted from the pre-state during the state transition (in case they are actually not deleted, they are ignored).
//
// The function uses post-state proofs to identify deleted keys and compute orphan nodes to add to
// the pre-state trie, ensuring it contains the required nodes for proper key deletion during the transition.
func (ns *ProvedNodeSet) AddOrphanNodes(postRoot gethcommon.Hash, postProofDB ethdb.KeyValueReader, keys ...[]byte) error {
	for _, key := range keys {
		// Verify the post-state proof
		v, prefix, node, err := VerifyProofWithLastNode(postRoot, key, postProofDB)
		if err != nil {
			return fmt.Errorf("invalid post-state proofs: %v", err)
		}

		if v != nil {
			// The key exists in the post-state, thus was not deleted, so we don't need to do anything
			continue
		}

		// The post-state proof is a valid exclusion proof
		// We now assess if the deletion of the key resulted in a trie reduction by checking the last proof node (post-state)
		sn, ok := mustDecodeNode(crypto.Keccak256(node), node).(*shortNode)
		if !ok {
			// The last proof node in the post-state is not a short node, this means that the deletion did not
			// result in any trie reduction, so there is no need to add orphan nodes to the pre-state trie
			continue
		}

		// The last proof node in the post-state is a short node, this means that the deletion resulted in a trie reduction
		// so we need to add orphan nodes to the pre-state trie

		// We add every possible shorter extension of the last node to the pre-state trie
		// This is overkill, but it's the easiest way to make sure that the necessary orphan is added to the pre-state trie
		// TODO: Optimize this by only adding the necessary orphan node
		shortNodes := reduceShortNode(sn)
		for subKey, orphan := range shortNodes {
			fullPath := prefix
			fullPath = append(fullPath, []byte(subKey)...)
			// If there is no node with the same key in the pre-state trie, we add the possible orphan node
			if ns.set.Nodes[string(fullPath)] == nil {
				ns.set.AddNode(fullPath, orphan)
			}
		}
	}

	return nil
}

func accountsProofDBAndKeys(accountProofs []*AccountProof) (ethdb.KeyValueReader, [][]byte, error) {
	keys := make([][]byte, 0)
	proofDB := memorydb.New()
	for _, accountProof := range accountProofs {
		// Create the trie key for the account
		keys = append(keys, AccountTrieKey(accountProof.Address))

		// Populate the proof database with the account proof
		err := FromHexProof(accountProof.Proof, proofDB)
		if err != nil {
			return nil, nil, err
		}
	}

	return proofDB, keys, nil
}

// AddAccountNodes adds the account nodes associated to the given account proofs to the node set
// For each account proof, it validates proof before adding the node to the set
func (ns *ProvedNodeSet) AddAccountNodes(accountRoot gethcommon.Hash, accountProofs []*AccountProof) error {
	proofDB, keys, err := accountsProofDBAndKeys(accountProofs)
	if err != nil {
		return nil
	}

	return ns.AddNodes(accountRoot, proofDB, keys...)
}

func (ns *ProvedNodeSet) AddAccountOrphanNodes(accountRoot gethcommon.Hash, accountProofs []*AccountProof) error {
	proofDB, keys, err := accountsProofDBAndKeys(accountProofs)
	if err != nil {
		return nil
	}

	return ns.AddOrphanNodes(accountRoot, proofDB, keys...)
}

// AddStorageNodes adds the storage nodes associated to the given storage proofs to the node set
// For each storage proof, it validates proof before adding the node to the set
func (ns *ProvedNodeSet) AddStorageNodes(storageRoot gethcommon.Hash, storageProofs []*StorageProof) error {
	proofDB, keys, err := storageProofDBAndKeys(storageProofs)
	if err != nil {
		return err
	}

	return ns.AddNodes(storageRoot, proofDB, keys...)
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
		err = FromHexProof(storageProof.Proof, proofDB)
		if err != nil {
			return nil, nil, err
		}
	}

	return proofDB, keys, nil
}

func (ns *ProvedNodeSet) AddStorageOrphanNodes(postRoot gethcommon.Hash, postProofs []*StorageProof) error {
	proofDB, keys, err := storageProofDBAndKeys(postProofs)
	if err != nil {
		return err
	}

	return ns.AddOrphanNodes(postRoot, proofDB, keys...)
}

func NodeSetFromProofs(owner gethcommon.Hash, root gethcommon.Hash, proofDB ethdb.KeyValueReader, keys ...[]byte) (*trienode.NodeSet, error) {
	provedNodeSet := NewProvedNodeSet(owner)
	err := provedNodeSet.AddNodes(root, proofDB, keys...)
	if err != nil {
		return nil, fmt.Errorf("failed to create node set from account state proofs for state root %v: %w", root.Hex(), err)
	}

	return provedNodeSet.Set(), nil
}

func NodeSetFromTransitionProof(owner, preRoot, postRoot gethcommon.Hash, preProofs, postProofs ethdb.KeyValueReader) (*trienode.NodeSet, error) {
	provedNodeSet := NewProvedNodeSet(owner)
	err := provedNodeSet.AddNodes(preRoot, preProofs)
	if err != nil {
		return nil, fmt.Errorf("failed to add account nodes for pre-state root %v: %w", preRoot.Hex(), err)
	}

	err = provedNodeSet.AddOrphanNodes(postRoot, postProofs)
	if err != nil {
		return nil, fmt.Errorf("failed to add account orphan nodes for post-state root %v: %w", postRoot.Hex(), err)
	}

	return provedNodeSet.Set(), nil
}

// NodeSetFromStateProofs constructs every state trie node set from a set of state proofs.
// It verifies every proof and an error is returned, if the proof is invalid.
func NodeSetFromStateProofs(stateRoot gethcommon.Hash, accountProofs []*AccountProof) (*trienode.MergedNodeSet, error) {
	// Compute node set for accounts world state
	provedAccountNodeSet := NewProvedNodeSet(AccountTrieOwner())
	err := provedAccountNodeSet.AddAccountNodes(stateRoot, accountProofs)
	if err != nil {
		return nil, fmt.Errorf("failed to create node set from account state proofs for state root %v: %w", stateRoot.Hex(), err)
	}

	mergedNodeSet := trienode.NewWithNodeSet(provedAccountNodeSet.Set())

	// Compute each account computed storage node set
	for _, accountProof := range accountProofs {
		if accountProof.StorageHash == (gethcommon.Hash{}) {
			continue
		}

		provedStorageNodeSet := NewProvedStorageNodeSet(accountProof.Address)
		err := provedStorageNodeSet.AddStorageNodes(
			accountProof.StorageHash,
			accountProof.Storage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create node set from storage proofs for account %v: %w", accountProof.Address.Hex(), err)
		}

		err = mergedNodeSet.Merge(provedStorageNodeSet.Set())
		if err != nil {
			return nil, fmt.Errorf("failed to merge storage node set for account %v: %w", accountProof.Address.Hex(), err)
		}
	}

	return mergedNodeSet, nil
}

func NodeSetFromStateTransitionProofs(preRoot, postRoot gethcommon.Hash, preProofs, postProofs []*AccountProof) (*trienode.MergedNodeSet, error) {
	// Compute node set for accounts world state
	provedAccountNodeSet := NewProvedNodeSet(AccountTrieOwner())
	err := provedAccountNodeSet.AddAccountNodes(preRoot, preProofs)
	if err != nil {
		return nil, fmt.Errorf("failed to add account nodes for pre-state root %v: %w", preRoot.Hex(), err)
	}

	err = provedAccountNodeSet.AddAccountOrphanNodes(postRoot, postProofs)
	if err != nil {
		return nil, fmt.Errorf("failed to add account orphan nodes for post-state root %v: %w", postRoot.Hex(), err)
	}

	mergedNodeSet := trienode.NewWithNodeSet(provedAccountNodeSet.Set())

	// Compute each account computed storage node set
	postProofsByAddress := make(map[gethcommon.Address]*AccountProof)
	for _, accountProof := range postProofs {
		postProofsByAddress[accountProof.Address] = accountProof
	}

	for _, accountProof := range preProofs {
		if accountProof.StorageHash == (gethcommon.Hash{}) {
			continue
		}
		provedStorageNodeSet := NewProvedStorageNodeSet(accountProof.Address)
		err := provedStorageNodeSet.AddStorageNodes(
			accountProof.StorageHash,
			accountProof.Storage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add storage nodes for account %v: %w", accountProof.Address.Hex(), err)
		}

		postAccountProof, ok := postProofsByAddress[accountProof.Address]
		if ok {
			err = provedStorageNodeSet.AddStorageOrphanNodes(
				postAccountProof.StorageHash,
				postAccountProof.Storage,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to add storage orphan nodes for account %v: %w", accountProof.Address.Hex(), err)
			}
		}

		err = mergedNodeSet.Merge(provedStorageNodeSet.Set())
		if err != nil {
			return nil, fmt.Errorf("failed to merge storage node set for account %v: %w", accountProof.Address.Hex(), err)
		}
	}

	return mergedNodeSet, nil
}

// reduceShortNode returns all possible short nodes that have the same child value as the original node but a shorter prefix key.

// e.g for a short node with node.Key: abcd and node.Value: 0x01, it returns:
// - abcd -> shortnode([], 0x01)
// - abc -> shortnode(d, 0x01)
// - ab -> shortnode(cd, 0x01)
// - a -> shortnode(bcd, 0x01)
func reduceShortNode(node *shortNode) map[string]*trienode.Node {
	shortNodes := make(map[string]*trienode.Node, 0)

	hasher := newHasher(false)
	defer returnHasherToPool(hasher)

	// If the node is a one-nibble node, it can not be reduced further
	if len(node.Key) == 1 {
		return shortNodes
	}

	for i := 1; i < len(node.Key); i++ {
		collapsed, hashed := hasher.proofHash(&shortNode{
			Key: node.Key[i:],
			Val: node.Val,
		})
		if hash, ok := hashed.(hashNode); ok {
			// If the node's database encoding is a hash (or is the
			// root node), it becomes a proof element.
			enc := nodeToBytes(collapsed)
			if !ok {
				hash = hasher.hashData(enc)
			}
			shortNodes[string(node.Key[:i])] = trienode.New(gethcommon.Hash(hash), enc)
		}
	}

	return shortNodes
}
