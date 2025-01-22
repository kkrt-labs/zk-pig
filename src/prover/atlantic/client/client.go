package atlantic

import (
	"context"
	"encoding/json"
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
	GetProof(ctx context.Context, atlanticQueryID string) (*Query, error)
}

// Layout represents the supported proof layout types
type Layout int

const (
	LayoutUnknown Layout = iota
	LayoutAuto
	LayoutRecursive
	LayoutRecursiveWithPoseidon
	LayoutSmall
	LayoutDex
	LayoutStarknet
	LayoutStarknetWithKeccak
	LayoutDynamic
)

var layouts = []string{
	"unknown",
	"auto",
	"recursive",
	"recursive_with_poseidon",
	"small",
	"dex",
	"starknet",
	"starknet_with_keccak",
	"dynamic",
}

func (l Layout) String() string {
	if l < LayoutUnknown || l > LayoutDynamic {
		return layouts[LayoutUnknown]
	}
	return layouts[l]
}

// MarshalJSON implements the json.Marshaler interface
func (l Layout) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (l *Layout) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for i, v := range layouts {
		if v == s {
			*l = Layout(i)
			return nil
		}
	}

	*l = LayoutUnknown
	return nil
}

// Prover represents the supported prover types
type Prover int

const (
	ProverUnknown Prover = iota
	ProverStarkwareSharp
)

var provers = []string{
	"stone",
	"starkware_sharp",
}

func (p Prover) String() string {
	if p < ProverUnknown || p > ProverStarkwareSharp {
		return provers[ProverUnknown]
	}
	return provers[p]
}

// MarshalJSON implements the json.Marshaler interface
func (p Prover) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (p *Prover) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for i, v := range provers {
		if v == s {
			*p = Prover(i)
			return nil
		}
	}

	*p = ProverUnknown
	return nil
}

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
	SharpQueries []Query `json:"sharpQueries"`
	Total        int     `json:"total"`
}

// Query represents a proof generation query
type Query struct {
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
