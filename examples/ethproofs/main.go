package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ethproofs "github.com/kkrt-labs/kakarot-controller/src/ethproofs/client"
	ethproofshttp "github.com/kkrt-labs/kakarot-controller/src/ethproofs/client/http"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("ETHPROOFS_API_KEY")
	if apiKey == "" {
		log.Fatal("ETHPROOFS_API_KEY environment variable is required")
	}

	// Create client
	client, err := ethproofshttp.NewClient(&ethproofshttp.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create a cluster
	cluster, err := client.CreateCluster(context.Background(), &ethproofs.CreateClusterRequest{
		Nickname:    "test-cluster",
		Description: "Test cluster for integration testing",
		Hardware:    "RISC-V Prover",
		CycleType:   "SP1",
		ProofType:   "Groth16",
		Configuration: []ethproofs.ClusterConfig{
			{
				InstanceType:  "t3.small",
				InstanceCount: 1,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create cluster: %v", err)
	}
	fmt.Printf("Created cluster with ID: %d\n", cluster.ID)

	// List all clusters
	clusters, err := client.ListClusters(context.Background())
	if err != nil {
		log.Fatalf("Failed to list clusters: %v", err)
	}
	fmt.Println("\nExisting clusters:")
	for _, c := range clusters {
		fmt.Printf("- %s (ID: %d)\n", c.Nickname, c.ID)
	}

	// Create a single machine
	machine, err := client.CreateMachine(context.Background(), &ethproofs.CreateMachineRequest{
		Nickname:     "test-machine",
		Description:  "Test single machine for integration testing",
		Hardware:     "RISC-V Prover",
		CycleType:    "SP1",
		ProofType:    "Groth16",
		InstanceType: "t3.small",
	})
	if err != nil {
		log.Fatalf("Failed to create machine: %v", err)
	}
	fmt.Printf("\nCreated machine with ID: %d\n", machine.ID)

	// List AWS pricing
	instances, err := client.ListAWSPricing(context.Background())
	if err != nil {
		log.Fatalf("Failed to list AWS pricing: %v", err)
	}
	fmt.Println("\nAvailable AWS instances:")
	for _, instance := range instances {
		fmt.Printf("- %s: $%.3f/hour (%d vCPUs, %.1fGB RAM)\n",
			instance.InstanceType,
			instance.HourlyPrice,
			instance.VCPU,
			instance.InstanceMemory)
	}

	// Demonstrate proof lifecycle
	// 1. Queue a proof
	queuedProof, err := client.QueueProof(context.Background(), &ethproofs.QueueProofRequest{
		BlockNumber: 12345,
		ClusterID:   cluster.ID,
	})
	if err != nil {
		log.Fatalf("Failed to queue proof: %v", err)
	}
	fmt.Printf("\nQueued proof with ID: %d\n", queuedProof.ProofID)

	// 2. Start proving
	startedProof, err := client.StartProving(context.Background(), &ethproofs.StartProvingRequest{
		BlockNumber: 12345,
		ClusterID:   cluster.ID,
	})
	if err != nil {
		log.Fatalf("Failed to start proving: %v", err)
	}
	fmt.Printf("Started proving proof with ID: %d\n", startedProof.ProofID)

	// 3. Submit completed proof
	provingCycles := int64(1000000)
	submittedProof, err := client.SubmitProof(context.Background(), &ethproofs.SubmitProofRequest{
		BlockNumber:   12345,
		ClusterID:     cluster.ID,
		ProvingTime:   60000, // 60 seconds in milliseconds
		ProvingCycles: &provingCycles,
		Proof:         "base64_encoded_proof_data_here",
		VerifierID:    "test-verifier",
	})
	if err != nil {
		log.Fatalf("Failed to submit proof: %v", err)
	}
	fmt.Printf("Submitted completed proof with ID: %d\n", submittedProof.ProofID)
}
