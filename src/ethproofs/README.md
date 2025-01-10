# EthProofs Client

The **EthProofs Client** is a Go package that provides a clean interface for interacting with the EthProofs API, enabling management of proving operations, clusters, and proof submissions.

## Overview

This package provides:
- A strongly-typed client interface for all EthProofs API endpoints
- HTTP implementation of the client interface
- Mock implementation for testing
- Comprehensive example usage

## Installation

```sh
go get github.com/kkrt-labs/kakarot-controller/src/ethproofs
```

## Authentication

All API endpoints require authentication using an API key. You can obtain one by:
1. Joining the EthProofs Telegram group
2. Requesting an API key from the team (contact Elias)
3. Using the API key in the environment variable `ETHPROOFS_API_KEY`
## Usage

### Creating a Client

```go
import (
    "github.com/kkrt-labs/kakarot-controller/src/ethproofs"
    ethproofshttp "github.com/kkrt-labs/kakarot-controller/src/ethproofs/http"
)

// Create client with API key
client, err := ethproofshttp.NewClient("your-api-key")
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

### AWS Pricing Information

```go
// List available AWS instances and pricing
instances, err := client.ListAWSPricing(context.Background())
for _, instance := range instances {
    fmt.Printf("- %s: $%.3f/hour (%d vCPUs, %.1fGB RAM)\n",
        instance.InstanceType,
        instance.HourlyPrice,
        instance.VCPU,
        instance.InstanceMemory)
}
```

## Testing

The package includes a mock client generated using `mockgen`. To use it in your tests:

```go
import (
    "testing"
    "github.com/golang/mock/gomock"
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

## API Documentation

For detailed API documentation, visit:
- [EthProofs API Documentation](https://staging--ethproofs.netlify.app/api.html)
- [EthProofs App Preview](https://staging--ethproofs.netlify.app/)
- [EthProofs Repository](https://github.com/ethproofs/ethproofs)

## Contributing

Interested in contributing? Check out our [Contributing Guidelines](../../CONTRIBUTING.md) to get started! 

# Testing EthProofs Client

## Prerequisites

1. Install required tools:
```bash
# Install mockgen
go install github.com/golang/mock/mockgen@latest

# Install test dependencies
go get github.com/stretchr/testify
```

## Generating Mocks

The mock client is automatically generated using mockgen. To regenerate the mocks:

```bash
# Generate mocks from the root directory
go generate ./src/ethproofs/...

# Or specifically for the client
go generate ./src/ethproofs/client.go
```

This will create/update `src/ethproofs/mock/client.go` based on the `Client` interface.

## Running Tests

### HTTP Client Tests
```bash
# Run all HTTP client tests
go test ./src/ethproofs/http/...

# Run specific test
go test ./src/ethproofs/http/... -run TestCreateCluster

# Run with verbose output
go test -v ./src/ethproofs/http/...

# Run with coverage
go test -cover ./src/ethproofs/http/...

# Generate coverage report
go test -coverprofile=coverage.out ./src/ethproofs/http/...
go tool cover -html=coverage.out
```

### Using Mock Client in Tests

Example of using the mock client in your tests:

```go
func TestYourFunction(t *testing.T) {
    // Create mock controller
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Create mock client
    mockClient := mock.NewMockClient(ctrl)

    // Set expectations
    mockClient.EXPECT().
        CreateCluster(gomock.Any(), &ethproofs.CreateClusterRequest{
            Nickname: "test-cluster",
        }).
        Return(&ethproofs.CreateClusterResponse{ID: 123}, nil)

    // Test your code that uses the client
    result, err := YourFunction(mockClient)
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

## Test Coverage

To view detailed test coverage:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./src/ethproofs/...

# View in browser
go tool cover -html=coverage.out

# View in terminal
go tool cover -func=coverage.out
```

## Continuous Integration

The tests are automatically run in CI/CD pipelines. Make sure all tests pass before submitting a PR:

```bash
# Run all tests and generate coverage
go test -race -coverprofile=coverage.out ./src/ethproofs/...

# Check coverage percentage
go tool cover -func=coverage.out | grep total:
```

## Directory Structure

```
src/ethproofs/
├── client.go           # Main interface definition
├── http/
│   ├── client.go      # HTTP implementation
│   ├── client_test.go # HTTP implementation tests
│   ├── clusters.go    # Clusters endpoint implementation
│   ├── proofs.go      # Proofs endpoint implementation
│   ├── machine.go     # Single machine endpoint implementation
│   └── aws.go         # AWS pricing endpoint implementation
└── mock/
    └── client.go      # Generated mock client
``` 