package main

import (
	"context"
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
	log.Printf("Generated proof with query ID: %s", proof.AtlanticQueryID)

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

	log.Printf("\nExisting proofs (showing %d of %d):", len(proofs.SharpQueries), proofs.Total)
	for _, p := range proofs.SharpQueries {
		log.Printf("- ID: %s", p.ID)
		log.Printf("  Status: %s", p.Status)
		log.Printf("  Created: %s", p.CreatedAt.Format(time.RFC3339))
		if p.CompletedAt != nil {
			log.Printf("  Completed: %s", p.CompletedAt.Format(time.RFC3339))
		}
		log.Printf("  Layout: %s", p.Layout)
		log.Printf("  Prover: %s", p.Prover)
		if p.GasUsed > 0 {
			log.Printf("  Gas Used: %d", p.GasUsed)
		}
		log.Print("\n")
	}

	// Get details of a specific proof
	queryID := proof.AtlanticQueryID // Using the ID from our generated proof
	proofDetails, err := client.GetProof(context.Background(), queryID)
	if err != nil {
		log.Fatalf("Failed to get proof details: %v", err)
	}

	log.Printf("\nDetailed proof information for %s:", queryID)
	log.Printf("Status: %s", proofDetails.Status)
	log.Printf("Program Hash: %s", proofDetails.ProgramHash)
	log.Printf("Program Fact Hash: %s", proofDetails.ProgramFactHash)
	if proofDetails.Price != "" {
		log.Printf("Price: %s", proofDetails.Price)
	}
	if proofDetails.GasUsed > 0 {
		log.Printf("Gas Used: %d", proofDetails.GasUsed)
	}
	if proofDetails.CreditsUsed > 0 {
		log.Printf("Credits Used: %d", proofDetails.CreditsUsed)
	}
	if len(proofDetails.Steps) > 0 {
		log.Printf("Steps: %v", proofDetails.Steps)
	}
}
