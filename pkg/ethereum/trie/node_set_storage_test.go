package trie

import (
	"encoding/json"
	"fmt"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type diffData struct {
	Pre  *StorageProof `json:"pre"`
	Post *StorageProof `json:"post"`
}

func loadStorageDiffData(t *testing.T) []*diffData {
	f, err := openTestData("storage_diff_proofs.json")
	require.NoError(t, err)

	var data []*diffData
	err = json.NewDecoder(f).Decode(&data)
	require.NoError(t, err)

	return data
}

func TestStorageNodeSet(t *testing.T) {
	diffs := loadStorageDiffData(t)
	for i, diff := range diffs {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) { testNodeSetFromDiffProofs(t, diff) })
	}
}

func testNodeSetFromDiffProofs(t *testing.T, d *diffData) {
	// --- Test that deletion succeeds on a partial pre-state trie with the orphan nodes ---
	partialPreTrieDBWithOrphans := newTestTrieDB()
	psetWithOrphans := NewStorageNodeSet(gethcommon.HexToAddress("0xac1fd47ebbdc78882866875621d63b30df10075688e2525532dc3a9607a6a47d"))
	preRoot := crypto.Keccak256Hash(hexutil.MustDecode(d.Pre.Proof[0]))
	err := psetWithOrphans.AddStorageNodes(preRoot, []*StorageProof{d.Pre})
	require.NoError(t, err)

	// Add the orphan nodes
	postRoot := crypto.Keccak256Hash(hexutil.MustDecode(d.Post.Proof[0]))
	err = psetWithOrphans.AddStorageOrphanNodes(postRoot, []*StorageProof{d.Post})
	require.NoError(t, err)

	err = partialPreTrieDBWithOrphans.Update(preRoot, gethcommon.Hash{}, 0, trienode.NewWithNodeSet(psetWithOrphans.Set()), triedb.NewStateSet())
	require.NoError(t, err)

	partialPreTrieWithOrphans, err := trie.New(trie.TrieID(preRoot), partialPreTrieDBWithOrphans)
	require.NoError(t, err)

	// Deletion should not fail as the orphan nodes have been added
	key := crypto.Keccak256(hexutil.MustDecode(d.Pre.Key))
	err = partialPreTrieWithOrphans.Delete(key)
	assert.NoError(t, err)
}
