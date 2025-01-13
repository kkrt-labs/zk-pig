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

func TestListAWSPricing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/aws-pricing-list", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		instances := []ethproofs.AWSInstance{
			{
				ID:             1,
				InstanceType:   "t3.small",
				HourlyPrice:    0.5,
				InstanceMemory: 2.0,
				VCPU:           2,
			},
		}
		json.NewEncoder(w).Encode(instances)
	}))
	defer server.Close()

	client, err := NewClient(&Config{
		Addr:   server.URL,
		APIKey: "test-key",
	})
	require.NoError(t, err)

	resp, err := client.ListAWSPricing(context.Background())
	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, "t3.small", resp[0].InstanceType)
}
