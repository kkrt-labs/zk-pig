package trie

import (
	"encoding/json"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type accountsData struct {
	Root     gethcommon.Hash `json:"root"`
	Accounts []*AccountProof `json:"accounts"`
}

func loadAccountsData(file string) (*accountsData, error) {
	f, err := openTestData(file)
	if err != nil {
		return nil, err
	}

	var data accountsData
	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, err
}

func TestAccountsNodeSet(t *testing.T) {
	data, err := loadAccountsData("account_proofs.json")
	require.NoError(t, err)

	// Generate the node set
	pset := NewAccountNodeSet()
	err = pset.AddAccountNodes(data.Root, data.Accounts)
	require.NoError(t, err)

	// Verify that the nodes can be added to a trie
	trieDB := newTestTrieDB()
	t.Logf("Nodes.Len %v", len(pset.Set().Nodes))
	t.Logf("Leaf.Len %v", len(pset.Set().Leaves))

	err = trieDB.Update(data.Root, gethtypes.EmptyRootHash, 0, trienode.NewWithNodeSet(pset.Set()), triedb.NewStateSet())
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
