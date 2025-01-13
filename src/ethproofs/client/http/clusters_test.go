package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ethproofs "github.com/kkrt-labs/kakarot-controller/src/ethproofs/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCluster(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/clusters", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Return response
		resp := ethproofs.CreateClusterResponse{ID: 123}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		Addr:   server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	req := &ethproofs.CreateClusterRequest{
		Nickname:      "test-cluster",
		Configuration: []ethproofs.ClusterConfig{},
	}

	resp, err := client.CreateCluster(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(123), resp.ID)
}

func TestListClusters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/clusters", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		clusters := []ethproofs.Cluster{
			{
				ID:       1,
				Nickname: "test-cluster",
			},
		}
		json.NewEncoder(w).Encode(clusters)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		Addr:   server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	resp, err := client.ListClusters(context.Background())
	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, "test-cluster", resp[0].Nickname)
}
