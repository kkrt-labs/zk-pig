package input

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

/**
 * This function is used to normalize the ProverInput struct for comparison purposes.
 * It ensures that the struct is in a consistent format for comparison, especially
 * important for fields like Codes, PreState, and AccessList.
 */
func NormalizeProverInput(input *ProverInput) *ProverInput {
	if input == nil {
		return nil
	}

	normalized := &ProverInput{
		ChainConfig: input.ChainConfig, // Assuming this is comparable as-is
		Blocks:      input.Blocks,      // Assuming this is comparable as-is
	}

	if input.Witness != nil {
		normalized.Witness = &Witness{
			Ancestors: input.Witness.Ancestors, // Assuming this is ordered by block number already
		}

		// Normalize Witness Codes ([][]byte)
		if len(input.Witness.Codes) > 0 {
			normalized.Witness.Codes = make([][]byte, len(input.Witness.Codes))
			copy(normalized.Witness.Codes, input.Witness.Codes)
			sort.Slice(normalized.Witness.Codes, func(i, j int) bool {
				return bytes.Compare(normalized.Witness.Codes[i], normalized.Witness.Codes[j]) < 0
			})
		}

		// Normalize Witness State ([]string)
		if len(input.Witness.State) > 0 {
			normalized.Witness.State = make([][]byte, len(input.Witness.State))
			copy(normalized.Witness.State, input.Witness.State)
			sort.Slice(normalized.Witness.State, func(i, j int) bool {
				return bytes.Compare(normalized.Witness.State[i], normalized.Witness.State[j]) < 0
			})
		}
	}

	return normalized
}

// Helper function to compare two ProverInput
func CompareProverInput(a, b *ProverInput) bool {
	if a == nil || b == nil {
		return a == b
	}

	normalizedA := NormalizeProverInput(a)
	normalizedB := NormalizeProverInput(b)

	// Convert to JSON for deep comparison
	jsonA, err := json.Marshal(normalizedA)
	if err != nil {
		return false
	}

	jsonB, err := json.Marshal(normalizedB)
	if err != nil {
		return false
	}

	return bytes.Equal(jsonA, jsonB)
}

// Test helper function that provides more detailed comparison information
func CompareProverInputWithDiff(a, b *ProverInput) (equal bool, diff string) {
	normalizedA := NormalizeProverInput(a)
	normalizedB := NormalizeProverInput(b)

	jsonA, err := json.MarshalIndent(normalizedA, "", "  ")
	if err != nil {
		return false, fmt.Sprintf("Failed to marshal first input: %v", err)
	}

	jsonB, err := json.MarshalIndent(normalizedB, "", "  ")
	if err != nil {
		return false, fmt.Sprintf("Failed to marshal second input: %v", err)
	}

	if !bytes.Equal(jsonA, jsonB) {
		return false, fmt.Sprintf("Inputs differ:\nFirst:\n%s\n\nSecond:\n%s",
			string(jsonA), string(jsonB))
	}

	return true, ""
}
