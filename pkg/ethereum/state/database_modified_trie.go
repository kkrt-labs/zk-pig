// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package state

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"

	// "github.com/ethereum/go-ethereum/trie" // do not use original geth trie
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/trie" // use modified trie
)

const (
	// Number of codehash->size associations to keep.
	codeSizeCacheSize = 100000

	// Cache size granted for caching clean code.
	codeCacheSize = 64 * 1024 * 1024

	// Number of address->curve point associations to keep.
	pointCacheSize = 4096
)

// CachingDB is an implementation of Database interface. It leverages both trie and
// state snapshot to provide functionalities for state access. It's meant to be a
// long-live object and has a few caches inside for sharing between blocks.
type CachingDB struct {
	disk          ethdb.KeyValueStore
	triedb        *triedb.Database
	snap          *snapshot.Tree
	codeCache     *lru.SizeConstrainedCache[common.Hash, []byte]
	codeSizeCache *lru.Cache[common.Hash, int]
	pointCache    *utils.PointCache
}

// NewDatabase creates a state database with the provided data sources.
func NewModifiedTrieDatabase(triedb *triedb.Database, snap *snapshot.Tree) *CachingDB {
	return &CachingDB{
		disk:          triedb.Disk(),
		triedb:        triedb,
		snap:          snap,
		codeCache:     lru.NewSizeConstrainedCache[common.Hash, []byte](codeCacheSize),
		codeSizeCache: lru.NewCache[common.Hash, int](codeSizeCacheSize),
		pointCache:    utils.NewPointCache(pointCacheSize),
	}
}

// NewDatabaseForTesting is similar to NewDatabase, but it initializes the caching
// db by using an ephemeral memory db with default config for testing.
func NewDatabaseForTesting() *CachingDB {
	return NewModifiedTrieDatabase(triedb.NewDatabase(rawdb.NewMemoryDatabase(), nil), nil)
}

// Reader returns a state reader associated with the specified state root.
func (db *CachingDB) Reader(stateRoot common.Hash) (gethstate.Reader, error) {
	var readers []gethstate.Reader

	// Set up the state snapshot reader if available. This feature
	// is optional and may be partially useful if it's not fully
	// generated.
	if db.snap != nil {
		sr, err := newStateReader(stateRoot, db.snap)
		if err == nil {
			readers = append(readers, sr) // snap reader is optional
		}
	}
	// Set up the trie reader, which is expected to always be available
	// as the gatekeeper unless the state is corrupted.
	tr, err := newTrieReader(stateRoot, db.triedb, db.pointCache)
	if err != nil {
		return nil, err
	}
	readers = append(readers, tr)

	return newMultiReader(readers...)
}

// OpenTrie opens the main account trie at a specific root hash.
func (db *CachingDB) OpenTrie(root common.Hash) (gethstate.Trie, error) {
	tr, err := trie.NewStateTrie(trie.StateTrieID(root), db.triedb)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

// OpenStorageTrie opens the storage trie of an account.
func (db *CachingDB) OpenStorageTrie(stateRoot common.Hash, address common.Address, root common.Hash, self gethstate.Trie) (gethstate.Trie, error) {
	// In the verkle case, there is only one tree. But the two-tree structure
	// is hardcoded in the codebase. So we need to return the same trie in this
	// case.
	if db.triedb.IsVerkle() {
		return self, nil
	}
	tr, err := trie.NewStateTrie(trie.StorageTrieID(stateRoot, crypto.Keccak256Hash(address.Bytes()), root), db.triedb)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

// ContractCode retrieves a particular contract's code.
func (db *CachingDB) ContractCode(address common.Address, codeHash common.Hash) ([]byte, error) {
	code, _ := db.codeCache.Get(codeHash)
	if len(code) > 0 {
		return code, nil
	}
	code = rawdb.ReadCode(db.disk, codeHash)
	if len(code) > 0 {
		db.codeCache.Add(codeHash, code)
		db.codeSizeCache.Add(codeHash, len(code))
		return code, nil
	}
	return nil, errors.New("not found")
}

// ContractCodeWithPrefix retrieves a particular contract's code. If the
// code can't be found in the cache, then check the existence with **new**
// db scheme.
func (db *CachingDB) ContractCodeWithPrefix(address common.Address, codeHash common.Hash) ([]byte, error) {
	code, _ := db.codeCache.Get(codeHash)
	if len(code) > 0 {
		return code, nil
	}
	code = rawdb.ReadCodeWithPrefix(db.disk, codeHash)
	if len(code) > 0 {
		db.codeCache.Add(codeHash, code)
		db.codeSizeCache.Add(codeHash, len(code))
		return code, nil
	}
	return nil, errors.New("not found")
}

// ContractCodeSize retrieves a particular contracts code's size.
func (db *CachingDB) ContractCodeSize(addr common.Address, codeHash common.Hash) (int, error) {
	if cached, ok := db.codeSizeCache.Get(codeHash); ok {
		return cached, nil
	}
	code, err := db.ContractCode(addr, codeHash)
	return len(code), err
}

// TrieDB retrieves any intermediate trie-node caching layer.
func (db *CachingDB) TrieDB() *triedb.Database {
	return db.triedb
}

// PointCache returns the cache of evaluated curve points.
func (db *CachingDB) PointCache() *utils.PointCache {
	return db.pointCache
}

// Snapshot returns the underlying state snapshot.
func (db *CachingDB) Snapshot() *snapshot.Tree {
	return db.snap
}

// mustCopyTrie returns a deep-copied trie.
func mustCopyTrie(t gethstate.Trie) gethstate.Trie {
	switch t := t.(type) {
	case *trie.StateTrie:
		return t.Copy()
	default:
		panic(fmt.Errorf("unknown trie type %T", t))
	}
}
