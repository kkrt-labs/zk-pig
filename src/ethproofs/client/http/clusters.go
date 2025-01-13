package http

import (
	"context"
	"net/http"

	ethproofs "github.com/kkrt-labs/kakarot-controller/src/ethproofs/client"
)

func (c *Client) CreateCluster(ctx context.Context, req *ethproofs.CreateClusterRequest) (*ethproofs.CreateClusterResponse, error) {
	var resp ethproofs.CreateClusterResponse
	if err := c.do(ctx, http.MethodPost, "/clusters", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) ListClusters(ctx context.Context) ([]ethproofs.Cluster, error) {
	var resp []ethproofs.Cluster
	if err := c.do(ctx, http.MethodGet, "/clusters", nil, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
