package proto

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/stretchr/testify/assert"
)

func TestInput(t *testing.T) {
	var testCases = []struct {
		desc  string
		input *input.ProverInput
	}{
		{
			desc:  "nil input",
			input: nil,
		},
		{
			desc:  "empty input",
			input: &input.ProverInput{},
		},
		{
			desc: "input with all fields set",
			input: &input.ProverInput{
				Version:     "1",
				Blocks:      []*input.Block{},
				Witness:     &input.Witness{},
				ChainConfig: &params.ChainConfig{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			protoInput := ToProto(tc.input)
			inputFromProto := FromProto(protoInput)
			assert.Equal(t, tc.input, inputFromProto)
		})
	}
}

func TestWitness(t *testing.T) {
	var testCases = []struct {
		desc  string
		input *input.Witness
	}{
		{
			desc:  "nil witness",
			input: nil,
		},
		{
			desc:  "witness with all fields set",
			input: &input.Witness{},
		},
		{
			desc: "witness with nil fields",
			input: &input.Witness{
				Ancestors: nil,
				State:     nil,
				Codes:     nil,
			},
		},
		{
			desc: "witness with empty fields",
			input: &input.Witness{
				Ancestors: []*gethtypes.Header{},
				State:     []hexutil.Bytes{{0x01, 0x02, 0x03}},
				Codes:     []hexutil.Bytes{{0x04, 0x05, 0x06}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			protoWitness := WitnessToProto(tc.input)
			witnessFromProto := WitnessFromProto(protoWitness)
			assert.Equal(t, tc.input, witnessFromProto)
		})
	}
}
