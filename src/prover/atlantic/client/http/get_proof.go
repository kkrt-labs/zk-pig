package http

import (
	"context"
	"fmt"
	"net/http"

	atlantic "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client"
)

type atlanticQueryRespMsg struct {
	AtlanticQuery atlantic.Query `json:"atlanticQuery"`
}

func (c *Client) GetProof(ctx context.Context, atlanticQueryID string) (*atlantic.Query, error) {
	path := fmt.Sprintf("/v1/atlantic-query/%s", atlanticQueryID)
	httpReq, err := c.prepareRequest(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return nil, err
	}

	var resp atlanticQueryRespMsg
	if err := c.doRequest(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp.AtlanticQuery, nil
}
