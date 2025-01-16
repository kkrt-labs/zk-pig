package http

import (
	"context"
	"fmt"
	"net/http"

	atlantic "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client"
)

func (c *Client) GetProof(ctx context.Context, atlanticQueryId string) (*atlantic.AtlanticQuery, error) {
	path := fmt.Sprintf("/v1/atlantic-query/%s", atlanticQueryId)
	httpReq, err := c.prepareRequest(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return nil, err
	}

	var resp struct {
		AtlanticQuery atlantic.AtlanticQuery `json:"atlanticQuery"`
	}
	if err := c.doRequest(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp.AtlanticQuery, nil
}
