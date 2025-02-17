package trie

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTrieDB() *triedb.Database {
	return triedb.NewDatabase(rawdb.NewMemoryDatabase(), &triedb.Config{HashDB: &hashdb.Config{}})
}

func openTestData(file string) (*os.File, error) {
	return os.Open(filepath.Join("testdata", file))
}

func loadStateProofs(t *testing.T, stateRoot string) []*AccountProof {
	file, err := openTestData(fmt.Sprintf("state_proofs_%v.json", stateRoot))
	require.NoError(t, err)

	var proofs []*AccountProof
	err = json.NewDecoder(file).Decode(&proofs)
	require.NoError(t, err)

	return proofs
}

func TestNodeSetFromStateProofs(t *testing.T) {
	stateRoots := []string{
		"0xac1fd47ebbdc78882866875621d63b30df10075688e2525532dc3a9607a6a47d",
		"0x5eabbd77c4ccf0c31b9321349aadcfb93d09762553909ab1b4f8784648b997f4",
	}

	for _, stateRoot := range stateRoots {
		t.Run(stateRoot, func(t *testing.T) { testNodeSetFromStateProofs(t, gethcommon.HexToHash(stateRoot)) })
	}
}

func testNodeSetFromStateProofs(t *testing.T, stateRoot gethcommon.Hash) {
	// Load the state accounts from file
	stateProofs := loadStateProofs(t, stateRoot.Hex())

	// Create a merged node set from the state accounts
	set, err := NodeSetFromStateProofs(stateRoot, stateProofs)
	require.NoError(t, err)

	// Ensure the node set can be inserted into a trie database
	trieDB := newTestTrieDB()
	err = trieDB.Update(stateRoot, gethcommon.Hash{}, 0, set, triedb.NewStateSet())
	require.NoError(t, err)

	stateTrie, err := trie.NewStateTrie(trie.StateTrieID(stateRoot), trieDB)
	require.NoError(t, err)

	// Verifies that the state has been properly set
	for i, accountProof := range stateProofs {
		t.Run(fmt.Sprintf("#%v:%v", i, accountProof.Address.Hex()), func(t *testing.T) { testStoredAccount(t, trieDB, stateTrie, stateRoot, accountProof) })
	}
}

func testStoredAccount(t *testing.T, trieDB *triedb.Database, stateTrie gethstate.Trie, stateRoot gethcommon.Hash, accountProof *AccountProof) {
	storedAccount, err := stateTrie.GetAccount(accountProof.Address)
	require.NoError(t, err)

	if accountProof.Nonce != 0 {
		assert.Equal(t, accountProof.Balance.ToInt().String(), storedAccount.Balance.String(), "Balance mismatch")
		assert.Equal(t, accountProof.CodeHash, gethcommon.Hash(storedAccount.CodeHash), "CodeHash mismatch")
		assert.Equal(t, accountProof.Nonce, storedAccount.Nonce, "Nonce mismatch")
		assert.Equal(t, accountProof.StorageHash.Hex(), storedAccount.Root.Hex(), "Root mismatch")
	}

	if len(accountProof.Storage) != 0 {
		storageTrie, err := trie.NewStateTrie(trie.StorageTrieID(stateRoot, StorageTrieOwner(accountProof.Address), accountProof.StorageHash), trieDB)
		require.NoError(t, err)

		// Verify storage
		for _, storageProof := range accountProof.Storage {
			storedValue, err := storageTrie.GetStorage(accountProof.Address, hexutil.MustDecode(storageProof.Key))
			assert.NoError(t, err, "Unexpected storage error for key %v", storageProof.Key)
			assert.Equal(t, storageProof.Value.String(), hexutil.EncodeBig(new(big.Int).SetBytes(storedValue)), "Storage value mismatch for key %v", storageProof.Key)
		}
	}
}

type testTransitionProofsData struct {
	PreRoot    gethcommon.Hash `json:"preRoot"`
	PostRoot   gethcommon.Hash `json:"postRoot"`
	PreProofs  []*AccountProof `json:"preProofs"`
	PostProofs []*AccountProof `json:"postProofs"`
}

func loadStateTransitionProofs(t *testing.T, id string) *testTransitionProofsData {
	file, err := openTestData(fmt.Sprintf("state_transition_proofs_%v.json", id))
	require.NoError(t, err)

	var data testTransitionProofsData
	err = json.NewDecoder(file).Decode(&data)
	require.NoError(t, err)

	return &data
}

func TestNodeSetFromStateTransitionProofs(t *testing.T) {
	ids := []string{
		"1_21344154",
		"1_21372583",
		"1_21373630",
	}

	for _, id := range ids {
		t.Run(id, func(t *testing.T) { testNodeSetFromStateTransitionProofs(t, id) })
	}
}

func testNodeSetFromStateTransitionProofs(t *testing.T, id string) {
	data := loadStateTransitionProofs(t, id)

	// Create a merged node set from the state accounts
	set, err := NodeSetFromStateTransitionProofs(data.PreRoot, data.PostRoot, data.PreProofs, data.PostProofs)
	require.NoError(t, err)

	// Ensure the node set can be inserted into a trie database
	trieDB := newTestTrieDB()
	err = trieDB.Update(data.PreRoot, gethcommon.Hash{}, 0, set, triedb.NewStateSet())
	require.NoError(t, err)

	// Create a custom trie
	stateTrie, err := trie.NewStateTrie(trie.StateTrieID(data.PreRoot), trieDB)
	require.NoError(t, err)

	// Verifies that the state has been properly created
	for i, accountProof := range data.PreProofs {
		t.Run(fmt.Sprintf("Creation#%v:%v", i, accountProof.Address.Hex()), func(t *testing.T) { testStoredAccount(t, trieDB, stateTrie, data.PreRoot, accountProof) })
	}

	// Test deletions can be performed
	postAccountByAddress := make(map[gethcommon.Address]*AccountProof)
	for _, account := range data.PostProofs {
		postAccountByAddress[account.Address] = account
	}

	for i, preAccountProof := range data.PreProofs {
		if postAccountProof, ok := postAccountByAddress[preAccountProof.Address]; ok {
			t.Run(fmt.Sprintf("Deletion#%v:%v", i, preAccountProof.Address.Hex()), func(t *testing.T) {
				testDeleteAccount(t, trieDB, stateTrie, data.PreRoot, preAccountProof, postAccountProof)
			})
		}
	}
}

func isAccountDeleted(accountProof *AccountProof) bool {
	return accountProof.Nonce == 0 && accountProof.Balance.ToInt().Sign() == 0 && accountProof.CodeHash == gethcommon.Hash{} && accountProof.StorageHash == gethcommon.Hash{}
}

func testDeleteAccount(t *testing.T, trieDB *triedb.Database, stateTrie gethstate.Trie, preStateRoot gethcommon.Hash, preAccountProof, postAccountProof *AccountProof) {
	if !isAccountDeleted(preAccountProof) && isAccountDeleted(postAccountProof) {
		// If the account has been deleted, delete it from the trie
		err := stateTrie.DeleteAccount(preAccountProof.Address)
		require.NoError(t, err, "Could not delete account %v", preAccountProof.Address)

		// Ensure the account has been deleted
		acc, err := stateTrie.GetAccount(preAccountProof.Address)
		assert.Error(t, err)
		assert.Nil(t, acc)
	}

	// Ensure the account storage has been deleted

	postStorageByKey := make(map[string]*StorageProof)
	for _, storageProof := range postAccountProof.Storage {
		postStorageByKey[storageProof.Key] = storageProof
	}

	for _, preStorageProof := range preAccountProof.Storage {
		if postProof, ok := postStorageByKey[preStorageProof.Key]; !(ok && postProof.Value.ToInt().Sign() == 0 && preStorageProof.Value.ToInt().Sign() != 0) {
			continue
		}
		// If the storage has been deleted, delete it from the trie
		storageTrie, err := trie.NewStateTrie(trie.StorageTrieID(preStateRoot, StorageTrieOwner(preAccountProof.Address), preAccountProof.StorageHash), trieDB)
		require.NoError(t, err)

		err = storageTrie.DeleteStorage(preAccountProof.Address, hexutil.MustDecode(preStorageProof.Key))
		require.NoError(t, err, "Could not delete storage for key %v", preStorageProof.Key)

		v, err := storageTrie.GetStorage(preAccountProof.Address, hexutil.MustDecode(preStorageProof.Key))
		assert.NoError(t, err, "Unexpected storage value for key %v (should have been deleted)", preStorageProof.Key)
		assert.Len(t, v, 0, "Unexpected storage value for key %v (should have been deleted)", preStorageProof.Key)
	}
}
