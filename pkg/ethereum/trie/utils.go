package trie

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// StorageTrieKey returns the key used to store a slot in storage trie.
func StorageTrieKey(slot []byte) []byte {
	return crypto.Keccak256(slot)
}

// AccountTrieKey returns the key used to store an account in the account trie.
func AccountTrieKey(addr gethcommon.Address) []byte {
	return crypto.Keccak256(addr.Bytes())
}

// StorageTrieOwner returns the owner of the storage trie for a given account.
func StorageTrieOwner(addr gethcommon.Address) gethcommon.Hash {
	return crypto.Keccak256Hash(addr.Bytes())
}

// AccountTrieOwner returns the owner of the account trie.
func AccountTrieOwner() gethcommon.Hash {
	return gethcommon.Hash{}
}
