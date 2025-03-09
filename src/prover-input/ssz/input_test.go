package ssz_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	ssz "github.com/kkrt-labs/zk-pig/src/prover-input/ssz"
)

func TestProverInputJSONToSSZ(t *testing.T) {

	proverInput := LoadProverInputJSON(t, "test-data.json")

	sszData, err := ssz.ToSSZ(proverInput)
	if err != nil {
		t.Fatalf("Failed to marshal ProverInput to SSZ: %v", err)
	}

	sszproverinput := &ssz.ProverInput{}
	if err := sszproverinput.UnmarshalSSZ(sszData); err != nil {
		t.Fatalf("Failed to unmarshal SSZ data: %v", err)
	}

	proverInputFromSSZ, err := ssz.ProverInputFromSSZ(sszproverinput)
	if err != nil {
		t.Fatalf("Failed to convert ProverInput from SSZ: %v", err)
	}

	jsonData, err := json.Marshal(proverInputFromSSZ)
	if err != nil {
		t.Fatalf("Failed to marshal ProverInput from SSZ to JSON: %v", err)
	}

	originalJSON, err := json.Marshal(proverInput)
	if err != nil {
		t.Fatalf("Failed to marshal original ProverInput to JSON: %v", err)
	}

	// Compare original and final JSON
	if !bytes.Equal(jsonData, originalJSON) {
		// store the original and final json in a file
		// this is for debugging purposes
		// TODO: remove this after fixing the issue
		// issue: the uncles and transactions field is not being serialized correctly
		// check the original.json and final.json files for more details
		os.WriteFile("original.json", originalJSON, 0644)
		os.WriteFile("final.json", jsonData, 0644)
		t.Errorf("Mismatch between original and final JSON")
	}
}

func LoadProverInputJSON(t *testing.T, path string) *input.ProverInput {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	var proverInput input.ProverInput
	if err := json.Unmarshal(jsonData, &proverInput); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	return &proverInput
}
