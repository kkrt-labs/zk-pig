package http

import (
	"context"
	"net/http"

	"github.com/kkrt-labs/kakarot-controller/src/ethproofs"
)

func (c *Client) CreateCluster(ctx context.Context, req *ethproofs.CreateClusterRequest) (*ethproofs.CreateClusterResponse, error) {
	httpReq, err := c.newRequest(ctx, http.MethodPost, "/clusters", req)
	if err != nil {
		return nil, err
	}

	var resp ethproofs.CreateClusterResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) ListClusters(ctx context.Context) ([]ethproofs.Cluster, error) {
	httpReq, err := c.newRequest(ctx, http.MethodGet, "/clusters", nil)
	if err != nil {
		return nil, err
	}

	var resp []ethproofs.Cluster
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
