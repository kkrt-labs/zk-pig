package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	atlantic "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProof(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/v1/atlantic-query/test-query-id", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("apiKey"))

		// Return response
		createdAt := time.Date(2023, 11, 7, 5, 31, 56, 0, time.UTC)
		resp := struct {
			AtlanticQuery atlantic.Query `json:"atlanticQuery"`
		}{
			AtlanticQuery: atlantic.Query{
				ID:        "test-query-id",
				Status:    "RECEIVED",
				CreatedAt: createdAt,
			},
		}
		err := json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		Addr:   server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	resp, err := client.GetProof(context.Background(), "test-query-id")
	require.NoError(t, err)
	assert.Equal(t, "test-query-id", resp.ID)
	assert.Equal(t, "RECEIVED", resp.Status)
}
