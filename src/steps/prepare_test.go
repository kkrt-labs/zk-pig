package steps

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
			p, err := NewPreparer(
				WithDataInclude(IncludeAll),
			)
			require.NoError(t, err)
			result, err := p.Prepare(context.Background(), &testDataInputs.PreflightData)
			require.NoError(t, err)
			require.NotNil(t, result)
			equal := input.CompareProverInput(&testDataInputs.ProverInput, result)
			require.True(t, equal)
		})
	}
}

func testDataInputsPath(filename string) string {
	return "testdata/" + filename
}
