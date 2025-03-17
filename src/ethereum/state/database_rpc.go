package state

import (
	"context"
	"fmt"
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
	"github.com/holiman/uint256"
	"github.com/kkrt-labs/go-utils/ethereum/rpc"
)

// RPCDatabase is a gethstate.Database that reads the state from a remote RPC node.
type RPCDatabase struct {
	gethstate.Database

	remote                 rpc.Client
	stateRootToBlockNumber map[gethcommon.Hash]*big.Int
	currentBlockNumber     *big.Int

	ctx context.Context
}

// HackDatabase creates a new state database that reads the state from a remote RPC node.
func Hack(db gethstate.Database, remote rpc.Client) *RPCDatabase {
	return HackWithContext(context.TODO(), db, remote)
}

func HackWithContext(ctx context.Context, db gethstate.Database, remote rpc.Client) *RPCDatabase {
	return &RPCDatabase{
		Database:               db,
		remote:                 remote,
		stateRootToBlockNumber: make(map[gethcommon.Hash]*big.Int),
		ctx:                    ctx,
	}
}

// MarkBlock records a mapping from state root to the corresponding block number.
// This is necessary as the underlying RPC node expects parameters to be block numbers and not a state root.
func (db *RPCDatabase) MarkBlock(header *gethtypes.Header) {
	db.stateRootToBlockNumber[header.Root] = header.Number
	db.currentBlockNumber = header.Number
}

func (db *RPCDatabase) getBlockNumber(stateRoot gethcommon.Hash) (*big.Int, error) {
	if blockNumber, ok := db.stateRootToBlockNumber[stateRoot]; ok {
		return blockNumber, nil
	}
	return nil, fmt.Errorf("missing block for state root %s", stateRoot.Hex())
}

// Reader implements the gethstate.Database interface.
func (db *RPCDatabase) Reader(root gethcommon.Hash) (gethstate.Reader, error) {
	blockNumber, err := db.getBlockNumber(root)
	if err != nil {
		return nil, err
	}

	// This is the reader that reads from the remote node.
	return &rpcReader{
		remote:      db.remote,
		blockNumber: blockNumber,
		root:        root,
		ctx:         db.ctx,
	}, nil
}

// OpenTrie implements the gethstate.Database interface.
func (db *RPCDatabase) OpenTrie(root gethcommon.Hash) (gethstate.Trie, error) {
	if tr, err := db.Database.OpenTrie(root); err == nil {
		return tr, nil
	}
	// We return a no-op trie to avoid some errors on block execution.
	// But it should be treated as suched and not used for any state access.
	return &NoOpTrie{}, nil
}

// OpenStorageTrie implements the gethstate.Database interface.
func (db *RPCDatabase) OpenStorageTrie(stateRoot gethcommon.Hash, address gethcommon.Address, root gethcommon.Hash, tr gethstate.Trie) (gethstate.Trie, error) {
	if tr, err := db.Database.OpenStorageTrie(stateRoot, address, root, tr); err == nil {
		return tr, nil
	}

	// We return a no-op trie to avoid some errors on block execution.
	// But it should be treated as suched and not used for any state access.
	return &NoOpTrie{}, nil
}

// rpcReader is a state reader that retrieves state information from a remote node (typically an archive node).
// it provides the ability to read account and storage information from a remote node
// it is useful when the local node does not have a full state trie
//
// Note:
//   - rpcReader needs a blockNumber and corresponding state root to retrieve the state information
//   - in case a block is not final and a re-org happens, the rpcReader may lead to inconsistent state information. If using
//     rpcReader on a non-finalized blocks, it is recommended to control the validity of the computed state information, with the finalized block
//     (in particular verify that the state root is the same as the one in the finalized block header)
type rpcReader struct {
	remote rpc.Client // Remote client to retrieve state information from remote node

	blockNumber *big.Int        // Block number to retrieve state information
	root        gethcommon.Hash // State root corresponding to the block number (it is assumed that the state root for the given block does not change (i.e. no re-org))

	ctx context.Context
}

// Account implementing Reader interface, retrieving the account associated with
// a particular address.
//
// - Returns a nil account if account is not available in the remote node
// - Returns an error only if remote node returns an error
// - The returned account is safe to modify after the call
func (r *rpcReader) Account(addr gethcommon.Address) (*gethtypes.StateAccount, error) {
	account, err := r.remote.GetProof(r.ctx, addr, nil, r.blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get proof for address %s and block %v: %v", addr.Hex(), r.blockNumber, err)
	}

	if account == nil {
		return nil, nil
	}

	balance, hasOverflowed := uint256.FromBig(account.Balance)
	if hasOverflowed {
		return nil, fmt.Errorf("failed to convert balance %v to uint256", account.Balance)
	}

	return &gethtypes.StateAccount{
		Nonce:    account.Nonce,
		Balance:  balance,
		Root:     account.StorageHash,
		CodeHash: account.CodeHash.Bytes(),
	}, nil
}

// Storage implementing Reader interface, retrieving the storage slot associated
// with a particular account address and slot key.
//
// - Returns an empty slot if it does not exist
// - Returns an error only if an unexpected issue occurs
// - The returned storage slot is safe to modify after the call
func (r *rpcReader) Storage(addr gethcommon.Address, slot gethcommon.Hash) (gethcommon.Hash, error) {
	value, err := r.remote.StorageAt(r.ctx, addr, slot, r.blockNumber)
	if err != nil {
		return gethcommon.Hash{}, fmt.Errorf("failed to get storage slot for address %s and slot %s and block %v: %v", addr.Hex(), slot.Hex(), r.blockNumber, err)
	}
	return gethcommon.BytesToHash(value), nil
}

func (r *rpcReader) Code(addr gethcommon.Address, codeHash gethcommon.Hash) ([]byte, error) {
	code, err := r.remote.CodeAt(r.ctx, addr, r.blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get code for address %s and block %v: %v", addr.Hex(), r.blockNumber, err)
	}
	return code, nil
}

func (r *rpcReader) CodeSize(addr gethcommon.Address, codeHash gethcommon.Hash) (int, error) {
	code, err := r.Code(addr, codeHash)
	return len(code), err
}

// Copy implementing Reader interface, returning a deep-copied state reader.
func (r *rpcReader) Copy() gethstate.Reader {
	return &rpcReader{
		blockNumber: r.blockNumber,
		remote:      r.remote,
		root:        r.root,
		ctx:         r.ctx,
	}
}

// NoOpTrie is a gethstate.Trie that does nothing.
type NoOpTrie struct{}

func (t *NoOpTrie) GetKey([]byte) []byte { return nil }
func (t *NoOpTrie) GetAccount(_ gethcommon.Address) (*gethtypes.StateAccount, error) {
	return nil, nil
}
func (t *NoOpTrie) GetStorage(_ gethcommon.Address, _ []byte) ([]byte, error) { return nil, nil }
func (t *NoOpTrie) UpdateAccount(_ gethcommon.Address, _ *gethtypes.StateAccount, _ int) error {
	return nil
}
func (t *NoOpTrie) UpdateStorage(_ gethcommon.Address, _, _ []byte) error { return nil }
func (t *NoOpTrie) DeleteAccount(_ gethcommon.Address) error              { return nil }
func (t *NoOpTrie) DeleteStorage(_ gethcommon.Address, _ []byte) error    { return nil }
func (t *NoOpTrie) UpdateContractCode(_ gethcommon.Address, _ gethcommon.Hash, _ []byte) error {
	return nil
}
func (t *NoOpTrie) Hash() gethcommon.Hash { return gethcommon.Hash{} }
func (t *NoOpTrie) Commit(_ bool) (gethcommon.Hash, *trienode.NodeSet) {
	return gethcommon.Hash{}, nil
}
func (t *NoOpTrie) Witness() map[string]struct{}                     { return nil }
func (t *NoOpTrie) NodeIterator(_ []byte) (trie.NodeIterator, error) { return nil, nil }
func (t *NoOpTrie) Prove(_ []byte, _ ethdb.KeyValueWriter) error     { return nil }
func (t *NoOpTrie) IsVerkle() bool                                   { return false }
