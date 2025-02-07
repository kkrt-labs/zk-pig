package ethereum

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
)

// WriteCodes fills an ethdb.Database with the provided bytecodes
func WriteCodes(db ethdb.Database, codes ...[]byte) {
	var (
		hasher = crypto.NewKeccakState()
		hash   = make([]byte, 32)
	)

	//nolint:errcheck // Can't fail
	for _, code := range codes {
		hasher.Reset()
		hasher.Write(code)
		hasher.Read(hash)

		rawdb.WriteCode(db, gethcommon.BytesToHash(hash), code)
	}
}

// WriteHeaders fills an ethdb.Database with the provided headers
func WriteHeaders(db ethdb.Database, headers ...*gethtypes.Header) {
	for _, header := range headers {
		rawdb.WriteHeader(db, header)
	}
}

// WriteNodesToHashDB fills an ethdb.Database with the provided nodes
func WriteNodesToHashDB(db ethdb.Database, nodes ...[]byte) {
	var (
		hasher = crypto.NewKeccakState()
		hash   = make([]byte, 32)
	)
	//nolint:errcheck // Can't fail
	for _, node := range nodes {
		hasher.Reset()
		hasher.Write(node)
		hasher.Read(hash)

		rawdb.WriteLegacyTrieNode(db, gethcommon.BytesToHash(hash), node)
	}
}
