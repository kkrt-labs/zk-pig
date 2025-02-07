package input

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
		Block:       input.Block,       // Assuming this is comparable as-is
		Ancestors:   input.Ancestors,   // Assuming this is ordered by block number already
	}

	// Normalize Codes ([][]byte)
	if len(input.Codes) > 0 {
		normalized.Codes = make([]hexutil.Bytes, len(input.Codes))
		copy(normalized.Codes, input.Codes)
		sort.Slice(normalized.Codes, func(i, j int) bool {
			return bytes.Compare(normalized.Codes[i], normalized.Codes[j]) < 0
		})
	}

	// Normalize PreState ([]string)
	if len(input.PreState) > 0 {
		normalized.PreState = make([]hexutil.Bytes, len(input.PreState))
		copy(normalized.PreState, input.PreState)
		sort.Slice(normalized.PreState, func(i, j int) bool {
			return bytes.Compare(normalized.PreState[i], normalized.PreState[j]) < 0
		})
	}

	// Normalize AccessList (map[gethcommon.Address][]string)
	if len(input.AccessList) > 0 {
		normalized.AccessList = make(map[gethcommon.Address][]hexutil.Bytes)

		// Get sorted addresses
		addresses := make([]gethcommon.Address, 0, len(input.AccessList))
		for addr := range input.AccessList {
			addresses = append(addresses, addr)
		}
		sort.Slice(addresses, func(i, j int) bool {
			return bytes.Compare(addresses[i].Bytes(), addresses[j].Bytes()) < 0
		})

		// Build normalized access list with sorted storage slots
		for _, addr := range addresses {
			if slots, ok := input.AccessList[addr]; ok {
				normalizedSlots := make([]hexutil.Bytes, len(slots))
				copy(normalizedSlots, slots)
				sort.Slice(normalizedSlots, func(i, j int) bool {
					return bytes.Compare(normalizedSlots[i], normalizedSlots[j]) < 0
				})
				normalized.AccessList[addr] = normalizedSlots
			}
		}
	}

	return normalized
}

// Helper function to compare two ProverInput
func CompareProverInputs(a, b *ProverInput) bool {
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
