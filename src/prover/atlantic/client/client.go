package atlantic

import (
	"context"
	"time"
)

// Package atlantic provides a Go client for the Herodotus Atlantic API.
//
// For more information about Atlantic, visit:
//   - API Documentation: https://docs.herodotus.cloud/atlantic/
//
//go:generate mockgen -source client.go -destination mock/client.go -package mock Client

// Client defines the interface for interacting with the Atlantic API
type Client interface {
	// Proofs
	GenerateProof(ctx context.Context, req *GenerateProofRequest) (*GenerateProofResponse, error)
	ListProofs(ctx context.Context, req *ListProofsRequest) (*ListProofsResponse, error)
	GetProof(ctx context.Context, atlanticQueryId string) (*AtlanticQuery, error)
}

// Layout represents the supported proof layout types
type Layout string

const (
	LayoutAuto                  Layout = "auto"
	LayoutRecursive             Layout = "recursive"
	LayoutRecursiveWithPoseidon Layout = "recursive_with_poseidon"
	LayoutSmall                 Layout = "small"
	LayoutDex                   Layout = "dex"
	LayoutStarknet              Layout = "starknet"
	LayoutStarknetWithKeccak    Layout = "starknet_with_keccak"
	LayoutDynamic               Layout = "dynamic"
)

// Prover represents the supported prover types
type Prover string

const (
	ProverStarkwareSharp Prover = "starkware_sharp"
)

// Request/Response types for Proof Generation
type GenerateProofRequest struct {
	PieFile []byte
	Layout  Layout
	Prover  Prover
}

type GenerateProofResponse struct {
	AtlanticQueryID string `json:"atlanticQueryId"`
}

// Request/Response types for Listing Proofs
type ListProofsRequest struct {
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

type ListProofsResponse struct {
	SharpQueries []AtlanticQuery `json:"sharpQueries"`
	Total        int             `json:"total"`
}

// AtlanticQuery represents a proof generation query
type AtlanticQuery struct {
	ID                string     `json:"id"`
	SubmittedByClient string     `json:"submittedByClient"`
	Status            string     `json:"status"`
	Step              string     `json:"step"`
	ProgramHash       string     `json:"programHash"`
	Layout            string     `json:"layout"`
	ProgramFactHash   string     `json:"programFactHash"`
	Price             string     `json:"price"`
	GasUsed           int64      `json:"gasUsed"`
	CreditsUsed       int64      `json:"creditsUsed"`
	TraceCreditsUsed  int64      `json:"traceCreditsUsed"`
	IsFactMocked      bool       `json:"isFactMocked"`
	Prover            Prover     `json:"prover"`
	Chain             string     `json:"chain"`
	Steps             []string   `json:"steps"`
	CreatedAt         time.Time  `json:"createdAt"`
	CompletedAt       *time.Time `json:"completedAt,omitempty"`
}
