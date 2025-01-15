package trie

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: write test for AccountState and StorageState

func newTestTrieDB() *triedb.Database {
	return triedb.NewDatabase(rawdb.NewMemoryDatabase(), &triedb.Config{HashDB: &hashdb.Config{}})
}

type accountsData struct {
	Root     gethcommon.Hash `json:"root"`
	Accounts []*accountData  `json:"accounts"`
}

type accountData struct {
	Address     gethcommon.Address `json:"address"`
	Proof       []string           `json:"accountProof"`
	Balance     hexutil.Big        `json:"balance"`
	Nonce       uint64             `json:"nonce"`
	CodeHash    gethcommon.Hash    `json:"codeHash"`
	StorageHash gethcommon.Hash    `json:"storageHash"`
}

func loadAccountsData(t *testing.T) *accountsData {
	f, err := os.Open("testdata/account_proofs.json")
	require.NoError(t, err)

	var data accountsData
	err = json.NewDecoder(f).Decode(&data)
	require.NoError(t, err)

	return &data
}

func TestNodeSetFromProof(t *testing.T) {
	t.Run("AddNodes#Accounts", func(t *testing.T) {
		data := loadAccountsData(t)
		testAccountsNodeSet(t, data)
	})

	// TODO: Add tests for AddNodes#Storage
}

func testAccountsNodeSet(t *testing.T, data *accountsData) {
	// Create a proofDB compatible with trie.VerifyProof
	proofDB := memorydb.New()
	for _, account := range data.Accounts {
		err := FromHexProof(account.Proof, proofDB)
		require.NoError(t, err)
	}

	keys := make([][]byte, len(data.Accounts))
	for i, account := range data.Accounts {
		keys[i] = crypto.Keccak256(account.Address.Bytes())
	}

	// Generate the node set
	pset := NewProvedAccountNodeSet()
	err := pset.AddNodes(data.Root, proofDB, keys...)
	require.NoError(t, err)

	// Verify that the nodes can be added to a trie
	trieDB := newTestTrieDB()
	t.Logf("Nodes.Len %v", len(pset.Set().Nodes))
	t.Logf("Leaf.Len %v", len(pset.Set().Leaves))

	err = trieDB.Update(data.Root, gethtypes.EmptyRootHash, 0, trienode.NewWithNodeSet(pset.Set()), nil)
	require.NoError(t, err)

	tr, err := trie.NewStateTrie(trie.StateTrieID(data.Root), trieDB)
	require.NoError(t, err)

	for _, account := range data.Accounts {
		t.Run(account.Address.String(), func(t *testing.T) {
			acc, err := tr.GetAccount(account.Address)
			assert.NoError(t, err)
			assert.Equal(t, account.Balance.ToInt(), acc.Balance.ToBig(), "Balance mismatch")
			assert.Equal(t, account.Nonce, acc.Nonce, "Nonce mismatch")
			assert.Equal(t, account.CodeHash.Bytes(), acc.CodeHash, "Code hash mismatch")
			assert.Equal(t, account.StorageHash, acc.Root, "Storage hash mismatch")
		})
	}
}

type data struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
	Proof []string `json:"proof"`
}

type diffData struct {
	Pre  data `json:"pre"`
	Post data `json:"post"`
}

func loadStorageDiffData(t *testing.T) []*diffData {
	f, err := os.Open("testdata/storage_diff_proofs.json")
	require.NoError(t, err)

	var data []*diffData
	err = json.NewDecoder(f).Decode(&data)
	require.NoError(t, err)

	return data
}

func TestNodeSetFromDiffProofs(t *testing.T) {
	t.Run("Storage", func(t *testing.T) {
		diffs := loadStorageDiffData(t)
		for i, diff := range diffs {
			t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) { testNodeSetFromDiffProofs(t, diff) })
		}
	})

	// TODO: Add tests for AccountNodeSet
}

func testNodeSetFromDiffProofs(t *testing.T, d *diffData) {
	// Generate a random owner (this has no impact on the test)
	owner := gethcommon.HexToHash("0xac1fd47ebbdc78882866875621d63b30df10075688e2525532dc3a9607a6a47d")

	// Key is the key to delete
	key := crypto.Keccak256(hexutil.MustDecode(d.Pre.Key))

	preRoot := crypto.Keccak256Hash(hexutil.MustDecode(d.Pre.Proof[0]))
	postRoot := crypto.Keccak256Hash(hexutil.MustDecode(d.Post.Proof[0]))

	preProofDB := memorydb.New()
	err := FromHexProof(d.Pre.Proof, preProofDB)
	require.NoError(t, err)

	postProofDB := memorydb.New()
	err = FromHexProof(d.Post.Proof, postProofDB)
	require.NoError(t, err)

	// --- Test that deletion succeeds on a partial pre-state trie with the orphan nodes ---
	partialPreTrieDBWithOrphans := newTestTrieDB()
	psetWithOrphans := NewProvedNodeSet(owner)
	err = psetWithOrphans.AddNodes(preRoot, preProofDB, key)
	require.NoError(t, err)

	// Add the orphan nodes
	err = psetWithOrphans.AddOrphanNodes(postRoot, postProofDB, key)
	require.NoError(t, err)

	err = partialPreTrieDBWithOrphans.Update(preRoot, gethcommon.Hash{}, 0, trienode.NewWithNodeSet(psetWithOrphans.Set()), nil)
	require.NoError(t, err)

	partialPreTrieWithOrphans, err := trie.New(trie.TrieID(preRoot), partialPreTrieDBWithOrphans)
	require.NoError(t, err)

	// Deletion should not fail as the orphan nodes have been added
	err = partialPreTrieWithOrphans.Delete(key)
	assert.NoError(t, err)
}

func loadStateProofs(t *testing.T, stateRoot string) []*AccountProof {
	file, err := os.Open(fmt.Sprintf("testdata/state_proofs_%v.json", stateRoot))
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
	err = trieDB.Update(stateRoot, gethcommon.Hash{}, 0, set, nil)
	require.NoError(t, err)

	stateTrie, err := trie.NewStateTrie(trie.StateTrieID(stateRoot), trieDB)
	require.NoError(t, err)

	// Verifies that the state has been properly set
	for i, acccountProof := range stateProofs {
		t.Run(fmt.Sprintf("#%v:%v", i, acccountProof.Address.Hex()), func(t *testing.T) { testStoredAccount(t, trieDB, stateTrie, stateRoot, acccountProof) })
	}
}

func testStoredAccount(t *testing.T, trieDB *triedb.Database, stateTrie gethstate.Trie, stateRoot gethcommon.Hash, acccountProof *AccountProof) {
	storedAccount, err := stateTrie.GetAccount(acccountProof.Address)
	require.NoError(t, err)

	if acccountProof.Nonce != 0 {
		assert.Equal(t, acccountProof.Balance.ToInt().String(), storedAccount.Balance.String(), "Balance mismatch")
		assert.Equal(t, acccountProof.CodeHash, gethcommon.Hash(storedAccount.CodeHash), "CodeHash mismatch")
		assert.Equal(t, acccountProof.Nonce, storedAccount.Nonce, "Nonce mismatch")
		assert.Equal(t, acccountProof.StorageHash.Hex(), storedAccount.Root.Hex(), "Root mismatch")
	}

	if len(acccountProof.Storage) != 0 {
		storageTrie, err := trie.NewStateTrie(trie.StorageTrieID(stateRoot, StorageTrieOwner(acccountProof.Address), acccountProof.StorageHash), trieDB)
		require.NoError(t, err)

		// Verify storage
		for _, storageProof := range acccountProof.Storage {
			storedValue, err := storageTrie.GetStorage(acccountProof.Address, hexutil.MustDecode(storageProof.Key))
			assert.NoError(t, err, "Unexpected storage error for key %v", storageProof.Key)
			assert.Equal(t, storageProof.Value.String(), hexutil.EncodeBig(new(big.Int).SetBytes(storedValue)), "Storage value mismatch for key %v", storageProof.Key)
		}
	}
}

func TestReduceExtensionNode(t *testing.T) {
	b := hexutil.MustDecode("0xf8669d208794ba15c4844f9ec07fc828e0db8d9056cb0b8f7bad6a4b665d9da3b846f8440180a07c0f9bdf8569ebbf56ba759f159ca4c03adf6fbdafcd174f1a9cb935a24c4b91a05b83bdbcc56b2e630f2807bbadd2b0c21619108066b92a58de081261089e9ce5")
	originalNode := mustDecodeNode(crypto.Keccak256(b), b).(*shortNode)
	shortenNodes := reduceShortNode(originalNode)

	assert.Len(t, shortenNodes, len(originalNode.Key)-1)
	for key, n := range shortenNodes {
		i := len(key)
		assert.Equal(t, originalNode.Key[:i], []byte(key))
		node := mustDecodeNode(n.Hash[:], n.Blob).(*shortNode)
		assert.Equal(t, originalNode.Val, node.Val)
		assert.Equal(t, originalNode.Key[i:], node.Key)
	}
}

type testTransitionProofsData struct {
	PreRoot    gethcommon.Hash `json:"preRoot"`
	PostRoot   gethcommon.Hash `json:"postRoot"`
	PreProofs  []*AccountProof `json:"preProofs"`
	PostProofs []*AccountProof `json:"postProofs"`
}

func loadStateTransitionProofs(t *testing.T, id string) *testTransitionProofsData {
	file, err := os.Open(fmt.Sprintf("testdata/state_transition_proofs_%v.json", id))
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
	err = trieDB.Update(data.PreRoot, gethcommon.Hash{}, 0, set, nil)
	require.NoError(t, err)

	// Create a custom trie
	stateTrie, err := NewStateTrie(StateTrieID(data.PreRoot), trieDB)
	require.NoError(t, err)

	// Verifies that the state has been properly created
	for i, acccountProof := range data.PreProofs {
		t.Run(fmt.Sprintf("Creation#%v:%v", i, acccountProof.Address.Hex()), func(t *testing.T) { testStoredAccount(t, trieDB, stateTrie, data.PreRoot, acccountProof) })
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
		if postProof, ok := postStorageByKey[preStorageProof.Key]; ok && postProof.Value.ToInt().Sign() == 0 && preStorageProof.Value.ToInt().Sign() != 0 {
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
}

func TestCustomTrie(t *testing.T) {
	owner := gethcommon.HexToHash("0xac1fd47ebbdc78882866875621d63b30df10075688e2525532dc3a9607a6a47d")
	gethTrie, err := trie.New(trie.StorageTrieID(gethcommon.Hash{}, owner, gethcommon.Hash{}), newTestTrieDB())
	require.NoError(t, err)

	keys := [][]byte{
		hexutil.MustDecode("0xaa00000000000000000000000000000000000000000000000000000000000000"),
		hexutil.MustDecode("0xaa00000000000000000000000000000000000000000000000000000000000fea"),
		hexutil.MustDecode("0xaa00000000000000000000000000000000000000000000000000000000000feb"),
		hexutil.MustDecode("0xaa00000000000000000000000000000000000000000000000000000000000fff"),
	}

	values := [][]byte{
		hexutil.MustDecode("0xbb00000000000000000000000000000000000000000000000000000000000000cc"),
		hexutil.MustDecode("0xbb00000000000000000000000000000000000000000000000000000000000feacc"),
		hexutil.MustDecode("0xbb00000000000000000000000000000000000000000000000000000000000febcc"),
		hexutil.MustDecode("0xbb00000000000000000000000000000000000000000000000000000000000fffcc"),
	}

	// Populate the trie with some keys
	for i := range keys {
		gethTrie.Update(keys[i], values[i])
		v, err := gethTrie.Get(keys[i])
		require.NoError(t, err)
		assert.Equal(t, values[i], v)
	}

	// Get the proofs for the keys
	preProofDB := memorydb.New()
	for i := range keys {
		err := gethTrie.Prove(keys[i], preProofDB)
		require.NoError(t, err)
	}
	preRoot := gethTrie.Hash()

	// Delete entry in the trie resulting in reduction of the trie
	err = gethTrie.Delete(keys[3])
	require.NoError(t, err)

	postProofsDB := memorydb.New()
	for i := range keys {
		err = gethTrie.Prove(keys[i], postProofsDB)
		require.NoError(t, err)
	}
	postRoot := gethTrie.Hash()

	assert.NotEqual(t, preRoot, postRoot)

	// Create a partial node set with the deleted key
	set, err := NodeSetFromProofs(owner, preRoot, preProofDB, keys[3])
	require.NoError(t, err)

	// --- Test delete on partial trie built on partial Geth trie ---
	t.Run("Geth Trie", func(t *testing.T) {
		gethTrieDB := newTestTrieDB()
		err = gethTrieDB.Update(preRoot, gethcommon.Hash{}, 0, trienode.NewWithNodeSet(set), nil)
		require.NoError(t, err)

		partialGethTrie, err := trie.New(trie.TrieID(preRoot), gethTrieDB)
		require.NoError(t, err)

		// Deletion should as we are basing on Geth original trie
		err = partialGethTrie.Delete(keys[3])
		assert.Error(t, err)
	})

	// --- Test delete on partial trie built on custom trie ---
	t.Run("Custom Trie", func(t *testing.T) {
		customTrieDB := newTestTrieDB()
		err = customTrieDB.Update(preRoot, gethcommon.Hash{}, 0, trienode.NewWithNodeSet(set), nil)
		require.NoError(t, err)

		partialCustomTrie, err := New(TrieID(preRoot), customTrieDB)
		require.NoError(t, err)

		// Deletion should not fail as we are basing on a modified version of the trie
		err = partialCustomTrie.Delete(keys[3])
		assert.NoError(t, err)

		// Ensure the root hash is the same
		assert.Equal(t, postRoot, partialCustomTrie.Hash())
	})
}
