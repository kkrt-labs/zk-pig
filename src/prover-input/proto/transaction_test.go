package proto

import (
	"math/big"
	"testing"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func blob(data []byte) kzg4844.Blob {
	var blob kzg4844.Blob
	copy(blob[:], data)
	return blob
}

func commitment(data []byte) kzg4844.Commitment {
	var commitment kzg4844.Commitment
	copy(commitment[:], data)
	return commitment
}

func proof(data []byte) kzg4844.Proof {
	var proof kzg4844.Proof
	copy(proof[:], data)
	return proof
}

func TestTransaction(t *testing.T) {
	type testCase struct {
		desc string
		tx   *gethtypes.Transaction
	}

	testCases := []testCase{
		{
			desc: "nil transaction",
			tx:   nil,
		},
		{
			desc: "LegacyTransaction#NonEmpty",
			tx: gethtypes.NewTx(&gethtypes.LegacyTx{
				Nonce:    1,
				GasPrice: big.NewInt(1000000000000000000),
				Gas:      100000,
				To:       &gethcommon.Address{1},
				Value:    big.NewInt(2000000000000000000),
				Data:     []byte("test"),
				V:        big.NewInt(3),
				R:        big.NewInt(2),
				S:        big.NewInt(1),
			}),
		},
		{
			desc: "LegacyTransaction#Empty",
			tx:   gethtypes.NewTx(&gethtypes.LegacyTx{}),
		},
		{
			desc: "AccessListTransaction#NonEmpty",
			tx: gethtypes.NewTx(&gethtypes.AccessListTx{
				ChainID:  big.NewInt(1),
				Nonce:    1,
				GasPrice: big.NewInt(1000000000000000000),
				Gas:      100000,
				To:       &gethcommon.Address{1},
				Value:    big.NewInt(2000000000000000000),
				Data:     []byte("test"),
				V:        big.NewInt(3),
				R:        big.NewInt(2),
				S:        big.NewInt(1),
				AccessList: gethtypes.AccessList{
					{
						Address:     gethcommon.Address{1},
						StorageKeys: []gethcommon.Hash{gethcommon.HexToHash("0x123")},
					},
				},
			}),
		},
		{
			desc: "AccessListTransaction#Empty",
			tx:   gethtypes.NewTx(&gethtypes.AccessListTx{}),
		},
		{
			desc: "DynamicFeeTransaction#NonEmpty",
			tx: gethtypes.NewTx(&gethtypes.DynamicFeeTx{
				ChainID:   big.NewInt(1),
				Nonce:     1,
				GasTipCap: big.NewInt(1000000000000000000),
				GasFeeCap: big.NewInt(2000000000000000000),
				Gas:       100000,
				To:        &gethcommon.Address{1},
				Value:     big.NewInt(3000000000000000000),
				Data:      []byte("test"),
				AccessList: gethtypes.AccessList{
					{
						Address:     gethcommon.Address{1},
						StorageKeys: []gethcommon.Hash{gethcommon.HexToHash("0x123")},
					},
				},
				V: big.NewInt(3),
				R: big.NewInt(2),
				S: big.NewInt(1),
			}),
		},
		{
			desc: "DynamicFeeTransaction#Empty",
			tx:   gethtypes.NewTx(&gethtypes.DynamicFeeTx{}),
		},
		{
			desc: "BlobTransaction#NonEmpty",
			tx: gethtypes.NewTx(&gethtypes.BlobTx{
				ChainID:   uint256.NewInt(1),
				Nonce:     1,
				GasTipCap: uint256.NewInt(1000000000000000000),
				GasFeeCap: uint256.NewInt(2000000000000000000),
				Gas:       100000,
				To:        gethcommon.Address{1},
				Value:     uint256.NewInt(3000000000000000000),
				Data:      []byte("test"),
				AccessList: gethtypes.AccessList{
					{
						Address:     gethcommon.Address{1},
						StorageKeys: []gethcommon.Hash{gethcommon.HexToHash("0x123")},
					},
				},
				V:          uint256.NewInt(3),
				R:          uint256.NewInt(2),
				S:          uint256.NewInt(1),
				BlobFeeCap: uint256.NewInt(100000),
				BlobHashes: []gethcommon.Hash{gethcommon.HexToHash("0x234")},
				Sidecar: &gethtypes.BlobTxSidecar{
					Blobs:       []kzg4844.Blob{blob([]byte("blob1"))},
					Commitments: []kzg4844.Commitment{commitment([]byte("commitment1"))},
					Proofs:      []kzg4844.Proof{proof([]byte("proof1"))},
				},
			}),
		},
		{
			desc: "BlobTransaction#Empty",
			tx:   gethtypes.NewTx(&gethtypes.BlobTx{}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			if testCase.tx != nil {
				testCase.tx.SetTime(time.Unix(0, 0)) // this is necessary to make the test deterministic
			}
			protoTx := TransactionToProto(testCase.tx)
			fromProto := TransactionFromProto(protoTx)
			assert.Equal(t, testCase.tx, fromProto)
		})
	}
}
