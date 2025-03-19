package proto

import (
	"math/big"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/stretchr/testify/assert"
)

func TestExtra(t *testing.T) {
	var testCases = []struct {
		desc  string
		input *input.Extra
	}{
		{
			desc:  "nil extra",
			input: nil,
		},
		{
			desc: "extra with all fields set",
			input: &input.Extra{
				AccessList: gethtypes.AccessList{
					{
						Address:     gethcommon.HexToAddress("0x123"),
						StorageKeys: []gethcommon.Hash{gethcommon.HexToHash("0x456")},
					},
				},
				StateDiffs: []*input.StateDiff{
					{
						Address:     gethcommon.HexToAddress("0x123"),
						PreAccount:  &input.Account{Balance: big.NewInt(100)},
						PostAccount: &input.Account{Balance: big.NewInt(200)},
						Storage:     []*input.StorageDiff{},
					},
				},
				Committed: [][]byte{gethcommon.HexToHash("0x456").Bytes()},
			},
		},
		{
			desc: "extra with empty fields",
			input: &input.Extra{
				AccessList: []gethtypes.AccessTuple{},
				StateDiffs: []*input.StateDiff{},
				Committed:  [][]byte{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			protoExtra := ExtraToProto(tc.input)
			extraFromProto := ExtraFromProto(protoExtra)
			assert.Equal(t, tc.input, extraFromProto)
		})
	}
}

func TestStateDiff(t *testing.T) {
	var testCases = []struct {
		desc  string
		input *input.StateDiff
	}{
		{
			desc:  "nil state diff",
			input: nil,
		},
		{
			desc: "state diff with all fields set",
			input: &input.StateDiff{
				Address:     gethcommon.HexToAddress("0x123"),
				PreAccount:  &input.Account{Balance: big.NewInt(100)},
				PostAccount: &input.Account{Balance: big.NewInt(200)},
				Storage: []*input.StorageDiff{
					{
						Slot:      gethcommon.HexToHash("0x456"),
						PreValue:  gethcommon.HexToHash("0x789"),
						PostValue: gethcommon.HexToHash("0xabc"),
					},
				},
			},
		},
		{
			desc: "state diff with nil fields",
			input: &input.StateDiff{
				Address:     gethcommon.HexToAddress("0x123"),
				PreAccount:  nil,
				PostAccount: nil,
				Storage:     nil,
			},
		},
		{
			desc: "state diff with empty fields",
			input: &input.StateDiff{
				Address:     gethcommon.HexToAddress("0x123"),
				PreAccount:  &input.Account{},
				PostAccount: &input.Account{},
				Storage:     []*input.StorageDiff{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			protoStateDiff := StateDiffToProto(tc.input)
			stateDiffFromProto := StateDiffFromProto(protoStateDiff)
			assert.Equal(t, tc.input, stateDiffFromProto)
		})
	}
}

func TestAccount(t *testing.T) {
	var testCases = []struct {
		desc  string
		input *input.Account
	}{
		{
			desc:  "nil account",
			input: nil,
		},
		{
			desc: "account with all fields set",
			input: &input.Account{
				Balance:     big.NewInt(100),
				CodeHash:    gethcommon.HexToHash("0x123"),
				Nonce:       1,
				StorageHash: gethcommon.HexToHash("0x456"),
			},
		},
		{
			desc: "account with nil fields",
			input: &input.Account{
				Balance: nil,
			},
		},
		{
			desc: "account with empty fields",
			input: &input.Account{
				Balance:     big.NewInt(0),
				CodeHash:    gethcommon.Hash{},
				Nonce:       0,
				StorageHash: gethcommon.Hash{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			protoAccount := AccountToProto(tc.input)
			accountFromProto := AccountFromProto(protoAccount)
			assert.Equal(t, tc.input, accountFromProto)
		})
	}
}
