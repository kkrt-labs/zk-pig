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

func TestCreateMachine(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/single-machine", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		resp := ethproofs.CreateMachineResponse{ID: 456}
		err := json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		Addr:   server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	req := &ethproofs.CreateMachineRequest{
		Nickname:     "test-machine",
		InstanceType: "t3.small",
	}

	resp, err := client.CreateMachine(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(456), resp.ID)
}
