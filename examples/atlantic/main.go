package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	atlantic "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client"
	atlantichttp "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client/http"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("ATLANTIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ATLANTIC_API_KEY environment variable is required")
	}

	// Create client
	client, err := atlantichttp.NewClient(&atlantichttp.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Read pie file
	pieFile, err := os.ReadFile("examples/atlantic/fibonacci_pie.zip")
	if err != nil {
		log.Fatalf("Failed to read pie file: %v", err)
	}

	// Generate a proof
	proof, err := client.GenerateProof(context.Background(), &atlantic.GenerateProofRequest{
		PieFile: pieFile,
		Layout:  atlantic.LayoutAuto,
		Prover:  atlantic.ProverStarkwareSharp,
	})
	if err != nil {
		log.Fatalf("Failed to generate proof: %v", err)
	}
	fmt.Printf("Generated proof with query ID: %s\n", proof.AtlanticQueryID)

	// List existing proofs with pagination
	limit := 10
	offset := 0
	proofs, err := client.ListProofs(context.Background(), &atlantic.ListProofsRequest{
		Limit:  &limit,
		Offset: &offset,
	})
	if err != nil {
		log.Fatalf("Failed to list proofs: %v", err)
	}

	fmt.Printf("\nExisting proofs (showing %d of %d):\n", len(proofs.SharpQueries), proofs.Total)
	for _, p := range proofs.SharpQueries {
		fmt.Printf("- ID: %s\n", p.ID)
		fmt.Printf("  Status: %s\n", p.Status)
		fmt.Printf("  Created: %s\n", p.CreatedAt.Format(time.RFC3339))
		if p.CompletedAt != nil {
			fmt.Printf("  Completed: %s\n", p.CompletedAt.Format(time.RFC3339))
		}
		fmt.Printf("  Layout: %s\n", p.Layout)
		fmt.Printf("  Prover: %s\n", p.Prover)
		if p.GasUsed > 0 {
			fmt.Printf("  Gas Used: %d\n", p.GasUsed)
		}
		fmt.Println()
	}

	// Get details of a specific proof
	queryID := proof.AtlanticQueryID // Using the ID from our generated proof
	proofDetails, err := client.GetProof(context.Background(), queryID)
	if err != nil {
		log.Fatalf("Failed to get proof details: %v", err)
	}

	fmt.Printf("\nDetailed proof information for %s:\n", queryID)
	fmt.Printf("Status: %s\n", proofDetails.Status)
	fmt.Printf("Program Hash: %s\n", proofDetails.ProgramHash)
	fmt.Printf("Program Fact Hash: %s\n", proofDetails.ProgramFactHash)
	if proofDetails.Price != "" {
		fmt.Printf("Price: %s\n", proofDetails.Price)
	}
	if proofDetails.GasUsed > 0 {
		fmt.Printf("Gas Used: %d\n", proofDetails.GasUsed)
	}
	if proofDetails.CreditsUsed > 0 {
		fmt.Printf("Credits Used: %d\n", proofDetails.CreditsUsed)
	}
	if len(proofDetails.Steps) > 0 {
		fmt.Printf("Steps: %v\n", proofDetails.Steps)
	}
}
