package blockinputs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testcases = []string{
	"Ethereum_Mainnet_21630258.json",
	"Optimism_Mainnet_130679264.json",
}

func TestPreparer(t *testing.T) {
	for _, name := range testcases {
		t.Run(name, func(t *testing.T) {
			testDataInputs := loadTestDataInputs(t, testDataInputsPath(name))
			p := NewPreparer().(*preparer)
			result, err := p.Prepare(context.Background(), &testDataInputs.HeavyProverInputs)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, CompareProverInputs(result, &testDataInputs.ProverInputs))
		})
	}
}

func testDataInputsPath(filename string) string {
	return "testdata/" + filename
}
