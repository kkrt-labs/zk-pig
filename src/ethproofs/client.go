package ethproofs

import (
	"context"
)

// Package ethproofs provides a Go client for the EthProofs API.
//
// For more information about EthProofs, visit:
//   - API Documentation: https://staging--ethproofs.netlify.app/api.html
//   - App Preview: https://staging--ethproofs.netlify.app/
//   - Repository: https://github.com/ethproofs/ethproofs
//
//go:generate mockgen -source client.go -destination mock/client.go -package mock Client

// Client defines the interface for interacting with the EthProofs API
type Client interface {
	// Clusters
	CreateCluster(ctx context.Context, req *CreateClusterRequest) (*CreateClusterResponse, error)
	ListClusters(ctx context.Context) ([]Cluster, error)

	// Single Machine
	CreateMachine(ctx context.Context, req *CreateMachineRequest) (*CreateMachineResponse, error)

	// Proofs
	QueueProof(ctx context.Context, req *QueueProofRequest) (*ProofResponse, error)
	StartProving(ctx context.Context, req *StartProvingRequest) (*ProofResponse, error)
	SubmitProof(ctx context.Context, req *SubmitProofRequest) (*ProofResponse, error)

	// AWS Pricing
	ListAWSPricing(ctx context.Context) ([]AWSInstance, error)
}

// Request/Response types for Clusters
type CreateClusterRequest struct {
	Nickname      string          `json:"nickname"`
	Description   string          `json:"description,omitempty"`
	Hardware      string          `json:"hardware,omitempty"`
	CycleType     string          `json:"cycle_type,omitempty"`
	ProofType     string          `json:"proof_type,omitempty"`
	Configuration []ClusterConfig `json:"configuration"`
}

type ClusterConfig struct {
	InstanceType  string `json:"instance_type"`
	InstanceCount int64  `json:"instance_count"`
}

type CreateClusterResponse struct {
	ID int64 `json:"id"`
}

type ListClustersResponse []Cluster

type Cluster struct {
	ID                   int64           `json:"id"`
	Nickname             string          `json:"nickname"`
	Description          string          `json:"description"`
	Hardware             string          `json:"hardware"`
	CycleType            string          `json:"cycle_type"`
	ProofType            string          `json:"proof_type"`
	ClusterConfiguration []ClusterConfig `json:"cluster_configuration"`
}

// Request/Response types for Single Machine
type CreateMachineRequest struct {
	Nickname     string `json:"nickname"`
	Description  string `json:"description,omitempty"`
	Hardware     string `json:"hardware,omitempty"`
	CycleType    string `json:"cycle_type,omitempty"`
	ProofType    string `json:"proof_type,omitempty"`
	InstanceType string `json:"instance_type"`
}

type CreateMachineResponse struct {
	ID int64 `json:"id"`
}

// Request/Response types for Proofs
type QueueProofRequest struct {
	BlockNumber int64 `json:"block_number"`
	ClusterID   int64 `json:"cluster_id"`
}

type StartProvingRequest struct {
	BlockNumber int64 `json:"block_number"`
	ClusterID   int64 `json:"cluster_id"`
}

type SubmitProofRequest struct {
	BlockNumber   int64  `json:"block_number"`
	ClusterID     int64  `json:"cluster_id"`
	ProvingTime   int64  `json:"proving_time"`
	ProvingCycles *int64 `json:"proving_cycles,omitempty"`
	Proof         string `json:"proof"`
	VerifierID    string `json:"verifier_id,omitempty"`
}

type ProofResponse struct {
	ProofID int64 `json:"proof_id"`
}

// Request/Response types for AWS Pricing
type ListAWSPricingResponse = []AWSInstance

type AWSInstance struct {
	ID              int64   `json:"id"`
	InstanceType    string  `json:"instance_type"`
	Region          string  `json:"region"`
	HourlyPrice     float64 `json:"hourly_price"`
	InstanceMemory  float64 `json:"instance_memory"`
	VCPU            int64   `json:"vcpu"`
	InstanceStorage string  `json:"instance_storage"`
	CreatedAt       string  `json:"created_at"`
}
