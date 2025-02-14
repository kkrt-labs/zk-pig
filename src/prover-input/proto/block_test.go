package proto

import (
	"math/big"
	"testing"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/kkrt-labs/go-utils/common"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/stretchr/testify/assert"
)

func TestBlock(t *testing.T) {
	type testCase struct {
		desc  string
		block *input.Block
	}

	testCases := []testCase{
		{
			desc:  "nil block",
			block: nil,
		},
		{
			desc:  "empty block",
			block: &input.Block{},
		},
		{
			desc: "block with zero values",
			block: &input.Block{
				Header:       &gethtypes.Header{},
				Transactions: []*gethtypes.Transaction{},
				Uncles:       []*gethtypes.Header{},
				Withdrawals:  []*gethtypes.Withdrawal{},
			},
		},
		{
			desc: "block with non-zero values",
			block: &input.Block{
				Header: &gethtypes.Header{
					Number: big.NewInt(1),
				},
				Transactions: []*gethtypes.Transaction{
					gethtypes.NewTx(&gethtypes.LegacyTx{}),
				},
				Uncles: []*gethtypes.Header{
					{},
				},
				Withdrawals: []*gethtypes.Withdrawal{
					{},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.block != nil {
				for _, tx := range tc.block.Transactions {
					tx.SetTime(time.Unix(0, 0)) // this is necessary to make the test deterministic
				}
			}
			protoBlock := BlockToProto(tc.block)
			blockFromProto := BlockFromProto(protoBlock)
			assert.Equal(t, tc.block, blockFromProto)
		})
	}
}

func TestHeader(t *testing.T) {
	type testCase struct {
		desc   string
		header *gethtypes.Header
	}

	testCases := []testCase{
		{
			desc:   "nil header",
			header: nil,
		},
		{
			desc:   "empty header",
			header: &gethtypes.Header{},
		},
		{
			desc: "header with zeros",
			header: &gethtypes.Header{
				ParentHash:       gethcommon.Hash{},
				UncleHash:        gethcommon.Hash{},
				Coinbase:         gethcommon.Address{},
				Root:             gethcommon.Hash{},
				TxHash:           gethcommon.Hash{},
				ReceiptHash:      gethcommon.Hash{},
				Bloom:            gethtypes.Bloom{},
				Difficulty:       big.NewInt(0),
				Number:           big.NewInt(0),
				GasLimit:         0,
				GasUsed:          0,
				Time:             0,
				Extra:            []byte{},
				MixDigest:        gethcommon.Hash{},
				Nonce:            gethtypes.BlockNonce{},
				BaseFee:          big.NewInt(0),
				WithdrawalsHash:  common.Ptr(gethcommon.Hash{}),
				BlobGasUsed:      common.Ptr(uint64(0)),
				ExcessBlobGas:    common.Ptr(uint64(0)),
				ParentBeaconRoot: common.Ptr(gethcommon.Hash{}),
				RequestsHash:     common.Ptr(gethcommon.Hash{}),
			},
		},
		{
			desc: "header with non-zero values",
			header: &gethtypes.Header{
				ParentHash:       gethcommon.Hash{0x1},
				UncleHash:        gethcommon.Hash{0x2},
				Coinbase:         gethcommon.Address{0x3},
				Root:             gethcommon.Hash{0x4},
				TxHash:           gethcommon.Hash{0x5},
				ReceiptHash:      gethcommon.Hash{0x6},
				Bloom:            gethtypes.Bloom{0x7},
				Difficulty:       big.NewInt(1000),
				Number:           big.NewInt(1000),
				GasLimit:         1000,
				GasUsed:          1000,
				Time:             1000,
				Extra:            []byte{0x8},
				MixDigest:        gethcommon.Hash{0x9},
				Nonce:            gethtypes.BlockNonce{0xa},
				BaseFee:          big.NewInt(1000),
				WithdrawalsHash:  common.Ptr(gethcommon.Hash{0xb}),
				BlobGasUsed:      common.Ptr(uint64(1000)),
				ExcessBlobGas:    common.Ptr(uint64(1000)),
				ParentBeaconRoot: common.Ptr(gethcommon.Hash{0xc}),
				RequestsHash:     common.Ptr(gethcommon.Hash{0xd}),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			protoHeader := HeaderToProto(tc.header)
			headerFromProto := HeaderFromProto(protoHeader)
			assert.Equal(t, tc.header, headerFromProto)
		})
	}
}
