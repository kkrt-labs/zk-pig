package blockinputs

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// ExpectedData represents the structure of the JSON file
type TestDataInputs struct {
	HeavyProverInputs HeavyProverInputs `json:"heavyProverInputs"`
	ProverInputs      ProverInputs      `json:"proverInputs"`
}

func loadTestDataInputs(t *testing.T, path string) *TestDataInputs {
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	var data TestDataInputs
	err = json.NewDecoder(f).Decode(&data)
	require.NoError(t, err)

	return &data
}

func TestUnmarshal(t *testing.T) {
	_ = loadTestDataInputs(t, testDataInputsPath("Ethereum_Mainnet_21630258.json"))
}

// TODO: Add unit-tests for the preflight block execution
// It is probably possible to create a mock ethrpc.Client that uses some preloaded preflight data
