package http

import (
	"context"
	"fmt"
	"net/http"

	atlantic "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client"
)

func (c *Client) ListProofs(ctx context.Context, req *atlantic.ListProofsRequest) (*atlantic.ListProofsResponse, error) {
	httpReq, err := c.prepareRequest(ctx, http.MethodGet, "/v1/atlantic-queries", nil, "")
	if err != nil {
		return nil, err
	}

	q := httpReq.URL.Query()
	if req.Limit != nil {
		q.Add("limit", fmt.Sprintf("%d", *req.Limit))
	}
	if req.Offset != nil {
		q.Add("offset", fmt.Sprintf("%d", *req.Offset))
	}
	httpReq.URL.RawQuery = q.Encode()

	var resp atlantic.ListProofsResponse
	if err := c.doRequest(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
