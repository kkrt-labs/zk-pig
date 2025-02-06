package rpcdb

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	rpcmock "github.com/kkrt-labs/go-utils/ethereum/rpc/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Below var and methods are taken from Geth
var headerPrefix = []byte("h") // headerPrefix + num (uint64 big endian) + hash -> header

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// headerKey = headerPrefix + num (uint64 big endian) + hash
func headerKey(number uint64, hash gethcommon.Hash) []byte {
	return append(append(headerPrefix, encodeBlockNumber(number)...), hash.Bytes()...)
}

func TestDecodeHeaderNumberAndHash(t *testing.T) {
	tests := []struct {
		key            []byte
		expectedNumber uint64
		expectedHash   gethcommon.Hash
		expectedOk     bool
	}{
		{headerKey(0, gethcommon.Hash{}), 0, gethcommon.Hash{}, true},
		{headerKey(1, gethcommon.HexToHash("0x1")), 1, gethcommon.HexToHash("0x1"), true},
		{headerKey(123456789, gethcommon.HexToHash("0xb44fb4e949d0f78f87f79ee46428f23a2a5713ce6fc6e0beb3dda78c2ac1ea55")), 123456789, gethcommon.HexToHash("0xb44fb4e949d0f78f87f79ee46428f23a2a5713ce6fc6e0beb3dda78c2ac1ea55"), true},
		{headerKey(1234, gethcommon.HexToHash("0xb44fb4e949d0f78f87f79ee46428f23a2a5713ce6fc6e0beb3dda78c2ac1ea55")), 1234, gethcommon.HexToHash("0xb44fb4e949d0f78f87f79ee46428f23a2a5713ce6fc6e0beb3dda78c2ac1ea55"), true},
		{headerKey(0, gethcommon.HexToHash("0xb44fb4e949d0f78f87f79ee46428f23a2a5713ce6fc6e0beb3dda78c2ac1ea55")), 0, gethcommon.HexToHash("0xb44fb4e949d0f78f87f79ee46428f23a2a5713ce6fc6e0beb3dda78c2ac1ea55"), true},
		{headerKey(1234, gethcommon.Hash{}), 1234, gethcommon.Hash{}, true},
		{append(headerKey(0, gethcommon.HexToHash("0x1")), byte('n')), 0, gethcommon.Hash{}, false},
		{[]byte{'h', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 0, gethcommon.Hash{}, true},
		{[]byte{'H', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 0, gethcommon.Hash{}, false},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%v:key=%v", i, string(test.key)), func(t *testing.T) {
			num, hash, ok := decodeHeaderNumberAndHash(test.key)
			if !test.expectedOk {
				require.False(t, ok, "decodeHeaderNumberAndHash(%v) returned ok=true, want ok=false", test.key)
			} else {
				require.True(t, ok, "decodeHeaderNumberAndHash(%v) returned ok=false, want ok=true", test.key)
				assert.Equal(t, test.expectedNumber, num)
				assert.Equal(t, test.expectedHash, hash)
			}
		})
	}
}

func TestDatabase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCli := rpcmock.NewMockClient(ctrl)
	db := Hack(rawdb.NewMemoryDatabase(), mockCli)

	t.Run("Get Header", func(t *testing.T) {
		header := &gethtypes.Header{
			Number:     big.NewInt(1234),
			ParentHash: gethcommon.HexToHash("0xb44fb4e949d0f78f87f79ee46428f23a2a5713ce6fc6e0beb3dda78c2ac1ea55"),
		}

		mockCli.EXPECT().HeaderByHash(gomock.Any(), header.Hash()).Return(header, nil)
		b, err := db.Get(headerKey(1234, header.Hash()))
		require.NoError(t, err)
		expectedB, _ := rlp.EncodeToBytes(header)
		assert.Equal(t, hexutil.Encode(expectedB), hexutil.Encode(b))
	})

	t.Run("Get Non-Header", func(t *testing.T) {
		b, err := db.Get([]byte("key"))
		require.Error(t, err)
		assert.Nil(t, b)
	})
}
