package blockinputs

import (
	"context"
	"testing"

	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/stretchr/testify/require"
)

var testcases = []string{
	"Ethereum_Mainnet_21465322.json",
}

func TestPreparer(t *testing.T) {
	for _, name := range testcases {
		t.Run(name, func(t *testing.T) {
			testDataInputs := loadTestDataInputs(t, testDataInputsPath(name))
			p := NewPreparer().(*preparer)
			result, err := p.Prepare(context.Background(), &testDataInputs.HeavyProverInput)
			require.NoError(t, err)
			require.NotNil(t, result)
			equal := input.CompareProverInputs(&testDataInputs.ProverInput, result)
			require.True(t, equal)
		})
	}
}

func testDataInputsPath(filename string) string {
	return "testdata/" + filename
}
