package rpcdb

import (
	"context"
	"encoding/binary"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kkrt-labs/go-utils/ethereum/rpc"
)

// Database wraps an ethdb.Database and fetches missing headers from a remote RPC server.
type Database struct {
	ethdb.Database
	remote rpc.Client
	ctx    context.Context
}

// Hack returns a new Database that fetches missing headers from the remote RPC server.
func Hack(db ethdb.Database, remote rpc.Client) *Database {
	return HackWithContext(context.TODO(), db, remote)
}

func HackWithContext(ctx context.Context, db ethdb.Database, remote rpc.Client) *Database {
	return &Database{
		Database: db,
		remote:   remote,
		ctx:      ctx,
	}
}

// decodeHeaderNumberAndHash decodes the header number and hash given a Geth ethdb database key.
// It returns the header number, hash, and a boolean indicating if the key is a header key.
func decodeHeaderNumberAndHash(key []byte) (uint64, gethcommon.Hash, bool) {
	if string(key[0]) != "h" || len(key) != 41 {
		return 0, gethcommon.Hash{}, false
	}

	return binary.BigEndian.Uint64(key[1:9]), gethcommon.BytesToHash(key[9:41]), true
}

// Get retrieves the value for a key.
// It intercepts the key to check if it is a header key.
// - If the key is a header key, it fetches the header from the remote RPC server.
// - Otherwise, it calls the underlying ethdb.Database.Get method.
func (db *Database) Get(key []byte) ([]byte, error) {
	// Decode the header number and hash from the key
	_, hash, ok := decodeHeaderNumberAndHash(key)
	if !ok {
		return db.Database.Get(key)
	}

	// Fetch the header from the remote RPC server
	// Note: We use the context.TODO() because the ethdb.Database.Get method does not accept a context.
	header, err := db.remote.HeaderByHash(db.ctx, hash)
	if err != nil {
		return nil, err
	}

	// Encode the header to RLP
	b, err := rlp.EncodeToBytes(header)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Has checks if the database has a key.
func (db *Database) Has(key []byte) (bool, error) {
	if _, err := db.Get(key); err != nil {
		return false, nil
	}
	return true, nil
}
