package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kkrt-labs/kakarot-controller/src/ethproofs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueProof(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proofs/queued", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		resp := ethproofs.ProofResponse{ProofID: 789}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		Addr:   server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	req := &ethproofs.QueueProofRequest{
		BlockNumber: 12345,
		ClusterID:   123,
	}

	resp, err := client.QueueProof(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(789), resp.ProofID)
}

func TestStartProving(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proofs/proving", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		resp := ethproofs.ProofResponse{ProofID: 790}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		Addr:   server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	req := &ethproofs.StartProvingRequest{
		BlockNumber: 12345,
		ClusterID:   123,
	}

	resp, err := client.StartProving(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(790), resp.ProofID)
}

func TestSubmitProof(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/proofs/proved", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		resp := ethproofs.ProofResponse{ProofID: 791}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		Addr:   server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	provingCycles := int64(1000000)
	req := &ethproofs.SubmitProofRequest{
		BlockNumber:   12345,
		ClusterID:     123,
		ProvingTime:   60000,
		ProvingCycles: &provingCycles,
		Proof:         "base64_encoded_proof_data",
		VerifierID:    "test-verifier",
	}

	resp, err := client.SubmitProof(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(791), resp.ProofID)
}
