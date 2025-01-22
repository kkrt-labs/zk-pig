package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	atlantic "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateProof(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/v1/proof-generation", r.URL.Path)
		assert.Equal(t, "test-key", r.URL.Query().Get("apiKey"))
		assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data"))

		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)

		// Verify form fields
		assert.Equal(t, "auto", r.FormValue("layout"))
		assert.Equal(t, "starkware_sharp", r.FormValue("prover"))

		// Verify file
		file, _, err := r.FormFile("pieFile")
		require.NoError(t, err)
		defer file.Close()

		fileContent, err := io.ReadAll(file)
		require.NoError(t, err)
		assert.Equal(t, []byte("test pie file"), fileContent)

		// Return response
		resp := atlantic.GenerateProofResponse{AtlanticQueryID: "test-query-id"}
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		Addr:   server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	req := &atlantic.GenerateProofRequest{
		PieFile: []byte("test pie file"),
		Layout:  atlantic.LayoutAuto,
		Prover:  atlantic.ProverStarkwareSharp,
	}

	resp, err := client.GenerateProof(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "test-query-id", resp.AtlanticQueryID)
}
