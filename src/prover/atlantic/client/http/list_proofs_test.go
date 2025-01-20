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

func TestListProofs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/v1/atlantic-queries", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("apiKey"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		assert.Equal(t, "0", r.URL.Query().Get("offset"))

		// Return response
		createdAt := time.Date(2023, 11, 7, 5, 31, 56, 0, time.UTC)
		resp := atlantic.ListProofsResponse{
			SharpQueries: []atlantic.Query{
				{
					ID:        "test-query-id",
					Status:    "RECEIVED",
					CreatedAt: createdAt,
				},
			},
			Total: 1,
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

	limit := 10
	offset := 0
	req := &atlantic.ListProofsRequest{
		Limit:  &limit,
		Offset: &offset,
	}

	resp, err := client.ListProofs(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	assert.Len(t, resp.SharpQueries, 1)
	assert.Equal(t, "test-query-id", resp.SharpQueries[0].ID)
}
