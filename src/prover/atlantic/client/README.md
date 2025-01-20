# Atlantic Client

The **Atlantic Client** is a Go package that provides a clean interface for interacting with the Herodotus Atlantic API, enabling generation and management of STARK proofs from Cairo program traces.

## Overview

This package provides:
- A strongly-typed client interface for all Atlantic API endpoints
- HTTP implementation using Azure's autorest for robust HTTP client management
- Mock implementation for testing using uber-go/mock
- Comprehensive example usage and tests

## Installation

```sh
go get github.com/kkrt-labs/kakarot-controller/src/prover/atlantic
```

## Authentication

All API endpoints require authentication using an API key. You can obtain one by:
1. Visiting the [Herodotus Documentation](https://docs.herodotus.cloud/atlantic/)
2. Requesting an API key from the Herodotus team

## Usage

### Creating a Client

```go
import (
    atlantic "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client"
    atlantichttp "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client/http"
)

// Create client with configuration
client, err := atlantichttp.NewClient(&atlantichttp.Config{
    APIKey: "your-api-key",
})
if err != nil {
    log.Fatal(err)
}
```

### Generating Proofs

```go
// Read pie file
pieFile, err := os.ReadFile("path/to/your/proof.pie")
if err != nil {
    log.Fatal(err)
}

// Generate a proof
proof, err := client.GenerateProof(context.Background(), &atlantic.GenerateProofRequest{
    PieFile: pieFile,
    Layout:  atlantic.LayoutAuto,
    Prover:  atlantic.ProverStarkwareSharp,
})
```

### Listing Proofs

```go
// List proofs with pagination
limit := 10
offset := 0
proofs, err := client.ListProofs(context.Background(), &atlantic.ListProofsRequest{
    Limit:  &limit,
    Offset: &offset,
})
```

### Getting Proof Details

```go
// Get details of a specific proof
proofDetails, err := client.GetProof(context.Background(), "your-query-id")
if err != nil {
    log.Fatal(err)
}
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
go test ./src/prover/atlantic/...

# Run with verbose output
go test -v ./src/prover/atlantic/...

# Run with coverage
go test -cover ./src/prover/atlantic/...

# Generate coverage report
go test -coverprofile=coverage.out ./src/prover/atlantic/...
go tool cover -html=coverage.out

# Run specific tests
go test -v ./src/prover/atlantic/... -run TestGenerateProof
go test -v ./src/prover/atlantic/... -run TestListProofs
```

### Using Mock Client in Tests

```go
import (
    "testing"
    "go.uber.org/mock/gomock"
    "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client/mock"
)

func TestYourFunction(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockClient := mock.NewMockClient(ctrl)
    
    // Set up expectations
    mockClient.EXPECT().
        GenerateProof(gomock.Any(), gomock.Any()).
        Return(&atlantic.GenerateProofResponse{AtlanticQueryID: "test-id"}, nil)

    // Use mockClient in your tests
}
```

## Directory Structure

```
src/prover/atlantic/
└── client/
    ├── README.md           # This file
    ├── client.go          # Main interface definition
    ├── http/
    │   ├── client.go      # HTTP implementation using autorest
    │   ├── config.go      # Client configuration
    │   ├── generate_proof.go    # Proof generation endpoint
    │   ├── generate_proof_test.go
    │   ├── list_proofs.go      # List proofs endpoint
    │   ├── list_proofs_test.go
    │   ├── get_proof.go        # Get proof details endpoint
    │   └── get_proof_test.go
    └── mock/
        └── client.go      # Generated mock client
```

## API Documentation

For detailed API documentation, visit:
- [Atlantic API Documentation](https://docs.herodotus.cloud/atlantic/)
- [Herodotus Website](https://herodotus.cloud/)

## Supported Features

### Layouts
- `auto`: Automatic layout selection
- `recursive`: Recursive proof layout
- `recursive_with_poseidon`: Recursive with Poseidon hash
- `small`: Small proof layout
- `dex`: DEX-specific layout
- `starknet`: StarkNet layout
- `starknet_with_keccak`: StarkNet with Keccak
- `dynamic`: Dynamic layout

### Provers
- `starkware_sharp`: StarkWare SHARP prover

## Contributing

Interested in contributing? Check out our [Contributing Guidelines](../../CONTRIBUTING.md) to get started! 