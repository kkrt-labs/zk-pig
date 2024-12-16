package blockinputs

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func testLoadExecInputs(t *testing.T, path string) *HeavyProverInputs {
	f, err := os.Open(path)
	require.NoError(t, err)

	var inputs HeavyProverInputs
	require.NoError(t, json.NewDecoder(f).Decode(&inputs))
	return &inputs
}

func TestUnmarshal(t *testing.T) {
	_ = testLoadExecInputs(t, "testdata/21372637.json")
}

// TODO: Add unit-tests for the preflight block execution
// It is probably possible to create a mock ethrpc.Client that uses some preloaded preflight data
