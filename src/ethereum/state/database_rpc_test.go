package state

import (
	"math/big"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	rpcmock "github.com/kkrt-labs/go-utils/ethereum/rpc/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRPCDatabaseImplementsInterface(t *testing.T) {
	assert.Implements(t, (*gethstate.Database)(nil), new(RPCDatabase))
}

func TestRPCDatabase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	remote := rpcmock.NewMockClient(ctrl)

	db := Hack(nil, remote)

	// Prepare test data
	stateRoot := gethcommon.HexToHash("0x6f39539da0b571e36e04cdee1ef9273ce168644d63822352f3a18c0504220166")
	accountAddr := gethcommon.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7")
	accountResult := &gethclient.AccountResult{
		Address:     accountAddr,
		Balance:     new(big.Int).SetUint64(0x1),
		Nonce:       0x1,
		CodeHash:    gethcommon.HexToHash("0xb44fb4e949d0f78f87f79ee46428f23a2a5713ce6fc6e0beb3dda78c2ac1ea55"),
		StorageHash: gethcommon.HexToHash("0x65d17ccfe8328a42712f5dcd7a8827eebab3341e1a8bd6a4cb741495b83bd026"),
	}
	blockNumber := big.NewInt(15)

	db.MarkBlock(&gethtypes.Header{
		Root:   stateRoot,
		Number: blockNumber,
	})

	reader, err := db.Reader(stateRoot)
	require.NoError(t, err)

	t.Run("reader.GetProof", func(t *testing.T) {
		remote.EXPECT().
			GetProof(gomock.Any(), accountAddr, nil, blockNumber).Return(nil, nil).
			Return(accountResult, nil)

		stateAccount, err := reader.Account(accountAddr)
		require.NoError(t, err)
		assert.Equal(t, accountResult.Balance, stateAccount.Balance.ToBig(), "Balance mismatch")
		assert.Equal(t, accountResult.Nonce, stateAccount.Nonce, "Nonce mismatch")
		assert.Equal(t, accountResult.CodeHash.Bytes(), stateAccount.CodeHash, "CodeHash mismatch")
		assert.Equal(t, accountResult.StorageHash, stateAccount.Root, "Root mismatch")
	})

	t.Run("reader.StorageAt", func(t *testing.T) {
		slot := gethcommon.HexToHash("0x0fb6d5609c9edab75bf587ea7449e6e6940d6e3df1992a1bd96ca8b74ffd16fc")
		remote.EXPECT().
			StorageAt(gomock.Any(), accountAddr, slot, blockNumber).Return(nil, nil).
			Return(hexutil.MustDecode("0xabcd"), nil)

		storageValue, err := reader.Storage(accountAddr, slot)
		require.NoError(t, err)
		assert.Equal(t, "0x000000000000000000000000000000000000000000000000000000000000abcd", storageValue.Hex(), "Storage value mismatch")
	})

	t.Run("db.CodeAt", func(t *testing.T) {
		remote.EXPECT().
			CodeAt(gomock.Any(), accountAddr, blockNumber).
			Return(hexutil.MustDecode("0xabcdef0123456789"), nil)
		code, err := db.ContractCode(accountAddr, gethcommon.Hash{})
		require.NoError(t, err)
		assert.Equal(t, "0xabcdef0123456789", hexutil.Encode(code), "Code mismatch")
	})

	t.Run("db.ContractCodeSize", func(t *testing.T) {
		remote.EXPECT().
			CodeAt(gomock.Any(), accountAddr, blockNumber).
			Return(hexutil.MustDecode("0xabcdef0123456789"), nil)
		size, err := db.ContractCodeSize(accountAddr, gethcommon.Hash{})
		require.NoError(t, err)
		assert.Equal(t, 8, size, "Code Size mismatch")
	})

	t.Run("reader.Copy", func(t *testing.T) {
		readerCopy := reader.Copy()
		assert.Equal(t, reader, readerCopy, "Reader copy mismatch")
	})
}

func TestStateAccessTrackerDatabaseImplementsInterface(t *testing.T) {
	assert.Implements(t, (*gethstate.Database)(nil), new(AccessTrackerDatabase))
}

func TestStateAccessTrackerDatabase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	remote := rpcmock.NewMockClient(ctrl)
	trackerDb := NewAccessTrackerDatabase(nil, remote)
	assert.Implements(t, (*gethstate.Database)(nil), trackerDb)

	// Prepare test data
	stateRoot := gethcommon.HexToHash("0x0123456789abcdef000000000000000000000000000000000000000000000000")
	blockNumber := big.NewInt(42)

	trackerDb.MarkBlock(&gethtypes.Header{
		Root:   stateRoot,
		Number: blockNumber,
	})

	reader, err := trackerDb.Reader(stateRoot)
	require.NoError(t, err, "failed to get reader for state root")

	accountAddr := gethcommon.HexToAddress("0x123456789abcdef0000000000000000000000000")
	accountResult := &gethclient.AccountResult{
		Address:  accountAddr,
		Balance:  new(big.Int).SetUint64(123),
		Nonce:    0x2,
		CodeHash: gethcommon.HexToHash("0xfeedfacefeedfacefeedfacefeedfacefeedfacefeedfacefeedfacefeedface"),
	}

	t.Run("trackerDb.Reader().Account", func(t *testing.T) {
		remote.EXPECT().
			GetProof(gomock.Any(), accountAddr, nil, blockNumber).
			Return(accountResult, nil)

		stateAccount, err := reader.Account(accountAddr)
		require.NoError(t, err, "failed to get account from trackerDb")
		assert.Equal(t, accountResult.Balance, stateAccount.Balance.ToBig(), "Balance mismatch")
		assert.Equal(t, accountResult.Nonce, stateAccount.Nonce, "Nonce mismatch")
		assert.Equal(t, accountResult.CodeHash.Bytes(), stateAccount.CodeHash, "CodeHash mismatch")
	})

	t.Run("trackerDb.Reader().Storage", func(t *testing.T) {
		slot := gethcommon.HexToHash("0xabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabcabca")
		remote.EXPECT().
			StorageAt(gomock.Any(), accountAddr, slot, blockNumber).
			Return(hexutil.MustDecode("0xdeadbeef"), nil)

		storageVal, err := reader.Storage(accountAddr, slot)
		require.NoError(t, err, "failed to get storage value")
		assert.Equal(t, "0x00000000000000000000000000000000000000000000000000000000deadbeef", storageVal.Hex(), "Storage value mismatch")
	})
}
