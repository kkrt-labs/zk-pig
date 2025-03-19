package trie

import (
	"fmt"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

// NodeSetFromStateProofs constructs a MPT node set from a set of state proofs.
// It verifies every proof and an error is returned, if any of the proofs is invalid.
func NodeSetFromStateProofs(stateRoot gethcommon.Hash, accountProofs []*AccountProof) (*trienode.MergedNodeSet, error) {
	// Compute node set for accounts world state
	accountNodeSet := NewAccountNodeSet()
	err := accountNodeSet.AddAccountNodes(stateRoot, accountProofs)
	if err != nil {
		return nil, fmt.Errorf("failed to create node set from account state proofs for state root %v: %w", stateRoot.Hex(), err)
	}

	mergedNodeSet := trienode.NewWithNodeSet(accountNodeSet.Set())

	// Compute each account computed storage node set
	for _, accountProof := range accountProofs {
		if accountProof.StorageHash == (gethcommon.Hash{}) {
			continue
		}

		storageNodeSet := NewStorageNodeSet(accountProof.Address)
		err := storageNodeSet.AddStorageNodes(
			accountProof.StorageHash,
			accountProof.Storage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create node set from storage proofs for account %v: %w", accountProof.Address.Hex(), err)
		}

		err = mergedNodeSet.Merge(storageNodeSet.Set())
		if err != nil {
			return nil, fmt.Errorf("failed to merge storage node set for account %v: %w", accountProof.Address.Hex(), err)
		}
	}

	return mergedNodeSet, nil
}

// NodeSetFromStateTransitionProofs constructs a MPT node set from a set of state transition proofs.
// It verifies every proof and an error is returned, if any of the proofs is invalid.
func NodeSetFromStateTransitionProofs(preRoot, postRoot gethcommon.Hash, preProofs, postProofs []*AccountProof) (*trienode.MergedNodeSet, error) {
	// Compute node set for accounts world state
	accountNodeSet := NewAccountNodeSet()
	err := accountNodeSet.AddAccountNodes(preRoot, preProofs)
	if err != nil {
		return nil, fmt.Errorf("failed to add account nodes for pre-state root %v: %w", preRoot.Hex(), err)
	}

	// Add orphan nodes for accounts that were deleted during the state transition
	err = accountNodeSet.AddAccountOrphanNodes(postRoot, postProofs)
	if err != nil {
		return nil, fmt.Errorf("failed to add account orphan nodes for post-state root %v: %w", postRoot.Hex(), err)
	}

	mergedNodeSet := trienode.NewWithNodeSet(accountNodeSet.Set())

	// Compute each account computed storage node set
	postProofsByAddress := make(map[gethcommon.Address]*AccountProof)
	for _, accountProof := range postProofs {
		postProofsByAddress[accountProof.Address] = accountProof
	}

	for _, accountProof := range preProofs {
		// If the account has no storage, we skip it
		if accountProof.StorageHash == (gethcommon.Hash{}) {
			continue
		}

		// Compute the storage node set for the account
		provedStorageNodeSet := NewStorageNodeSet(accountProof.Address)
		err := provedStorageNodeSet.AddStorageNodes(
			accountProof.StorageHash,
			accountProof.Storage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add storage nodes for account %v: %w", accountProof.Address.Hex(), err)
		}

		// Add orphan nodes for storage that were deleted during the state transition
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

// AddNodes adds the nodes associated to the given keys to the node set
// For each key, it validates proof before adding the node to the set
// - root is the state root hash
// - proofDB is a database containing all the necessary trie nodes (keyed by node hash)
// - keys is a list of keys to insert into the node set
func AddNodes(set *trienode.NodeSet, root gethcommon.Hash, proofDB ethdb.KeyValueReader, keys ...[]byte) error {
	for _, key := range keys {
		proof, err := trie.VerifyProofWithProof(root, key, proofDB)
		if err != nil {
			return fmt.Errorf("failed to verify proof for key %x: %v", key, err)
		}

		for _, n := range proof.Nodes() {
			set.AddNode(n.Path, n.Node)
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
func AddOrphanNodes(set *trienode.NodeSet, postRoot gethcommon.Hash, postProofDB ethdb.KeyValueReader, keys ...[]byte) error {
	for _, key := range keys {
		// Verify the post-state proof
		// v, prefix, node, err := trie.VerifyProofWithLastNode(postRoot, key, postProofDB)
		// if err != nil {
		// 	return fmt.Errorf("invalid post-state proofs: %v", err)
		// }

		proof, err := trie.VerifyProofWithProof(postRoot, key, postProofDB)
		if err != nil {
			return fmt.Errorf("failed to verify proof for key %x: %v", key, err)
		}

		if proof.Value() != nil {
			// The key exists in the post-state, thus was not deleted, so we don't need to do anything
			continue
		}

		// The last proof node in the post-state is a short node, this means that the deletion resulted in a trie reduction
		// so we need to add orphan nodes to the pre-state trie

		// We add every possible shorter extension of the last node to the pre-state trie
		// This is overkill, but it's the easiest way to make sure that the necessary orphan is added to the pre-state trie
		// TODO: Optimize this by only adding the necessary orphan node
		lastNode := proof.Nodes()[len(proof.Nodes())-1]
		shortNodes, err := trie.ShortenShortNode(lastNode.Node.Blob)
		if err != nil {
			// The last proof node in the post-state is not a short node, this means that the deletion did not
			// result in any trie reduction, so there is no need to add orphan nodes to the pre-state trie
			continue
		}

		for _, sn := range shortNodes {
			fullPath := lastNode.Path
			fullPath = append(fullPath, sn.Path...)
			// If there is no node with the same key in the pre-state trie, we add the possible orphan node
			if set.Nodes[string(fullPath)] == nil {
				set.AddNode(fullPath, sn.Node)
			}
		}
	}

	return nil
}
