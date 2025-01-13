# EthProofs Client

The **EthProofs Client** is a Go package that provides a clean interface for interacting with the EthProofs API, enabling management of proving operations, clusters, and proof submissions.

## Overview

This package provides:
- A strongly-typed client interface for all EthProofs API endpoints
- HTTP implementation using Azure's autorest for robust HTTP client management
- Mock implementation for testing using uber-go/mock
- Comprehensive example usage and tests

## Installation

```sh
go get github.com/kkrt-labs/kakarot-controller/src/ethproofs
```

## Authentication

All API endpoints require authentication using an API key. You can obtain one by:
1. Joining the EthProofs Slack workspace
2. Requesting an API key from the team (contact Elias)

## Usage

### Creating a Client

```go
import (
    "github.com/kkrt-labs/kakarot-controller/src/ethproofs"
    ethproofshttp "github.com/kkrt-labs/kakarot-controller/src/ethproofs/http"
)

// Create client with configuration
client, err := ethproofshttp.NewClient(&ethproofshttp.Config{
    Addr:   "https://staging--ethproofs.netlify.app/api/v0",
    APIKey: "your-api-key",
})
if err != nil {
    log.Fatal(err)
}
```

### Managing Clusters

```go
// Create a cluster
cluster, err := client.CreateCluster(context.Background(), &ethproofs.CreateClusterRequest{
    Nickname:    "test-cluster",
    Description: "Test cluster for proving operations",
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

// List all clusters
clusters, err := client.ListClusters(context.Background())
```

### Managing Single Machines

```go
// Create a single machine
machine, err := client.CreateMachine(context.Background(), &ethproofs.CreateMachineRequest{
    Nickname:     "test-machine",
    Description:  "Single machine for proving",
    Hardware:     "RISC-V Prover",
    CycleType:    "SP1",
    ProofType:    "Groth16",
    InstanceType: "t3.small",
})
```

### Proof Lifecycle

```go
// 1. Queue a proof
queuedProof, err := client.QueueProof(context.Background(), &ethproofs.QueueProofRequest{
    BlockNumber: 12345,
    ClusterID:   cluster.ID,
})

// 2. Start proving
startedProof, err := client.StartProving(context.Background(), &ethproofs.StartProvingRequest{
    BlockNumber: 12345,
    ClusterID:   cluster.ID,
})

// 3. Submit completed proof
provingCycles := int64(1000000)
submittedProof, err := client.SubmitProof(context.Background(), &ethproofs.SubmitProofRequest{
    BlockNumber:    12345,
    ClusterID:     cluster.ID,
    ProvingTime:   60000, // milliseconds
    ProvingCycles: &provingCycles,
    Proof:         "base64_encoded_proof_data",
    VerifierID:    "test-verifier",
})
```

## Testing

### Prerequisites

1. Install required tools:
```bash
# Install uber-go/mock
make mockgen-install
```

### Running Tests

```bash
# Run all tests
go test ./src/ethproofs/...

# Run with verbose output
go test -v ./src/ethproofs/...

# Run with coverage
go test -cover ./src/ethproofs/...

# Generate coverage report
go test -coverprofile=coverage.out ./src/ethproofs/...
go tool cover -html=coverage.out

# Run specific domain tests
go test -v ./src/ethproofs/... -run TestCreateCluster
go test -v ./src/ethproofs/... -run TestListAWSPricing
```

### Using Mock Client in Tests

```go
import (
    "testing"
    "go.uber.org/mock/gomock"
    "github.com/kkrt-labs/kakarot-controller/src/ethproofs/mock"
)

func TestYourFunction(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockClient := mock.NewMockClient(ctrl)
    
    // Set up expectations
    mockClient.EXPECT().
        CreateCluster(gomock.Any(), gomock.Any()).
        Return(&ethproofs.CreateClusterResponse{ID: 123}, nil)

    // Use mockClient in your tests
}
```

## Directory Structure

```
src/ethproofs/
├── client.go           # Main interface definition
├── http/
│   ├── client.go      # HTTP implementation using autorest
│   ├── config.go      # Client configuration
│   ├── clusters.go    # Clusters endpoint implementation
│   ├── clusters_test.go
│   ├── proofs.go      # Proofs endpoint implementation
│   ├── proofs_test.go
│   ├── machine.go     # Single machine endpoint implementation
│   ├── machine_test.go
│   ├── aws.go         # AWS pricing endpoint implementation
│   └── aws_test.go
└── mock/
    └── client.go      # Generated mock client
```

## API Documentation

For detailed API documentation, visit:
- [EthProofs API Documentation](https://staging--ethproofs.netlify.app/api.html)
- [EthProofs App Preview](https://staging--ethproofs.netlify.app/)
- [EthProofs Repository](https://github.com/ethproofs/ethproofs)

## Contributing

Interested in contributing? Check out our [Contributing Guidelines](../../CONTRIBUTING.md) to get started! 