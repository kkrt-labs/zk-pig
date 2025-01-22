package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	atlantic "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client"
)

func (c *Client) ListProofs(ctx context.Context, req *atlantic.ListProofsRequest) (*atlantic.ListProofsResponse, error) {
	params := map[string]interface{}{
		"apiKey": c.cfg.APIKey,
	}
	if req.Limit != nil {
		params["limit"] = fmt.Sprintf("%d", *req.Limit)
	}
	if req.Offset != nil {
		params["offset"] = fmt.Sprintf("%d", *req.Offset)
	}

	httpReq, err := autorest.CreatePreparer(
		autorest.WithMethod(http.MethodGet),
		autorest.WithPath("/v1/atlantic-queries"),
		autorest.WithQueryParameters(params),
	).Prepare(newRequest(ctx))
	if err != nil {
		return nil, err
	}

	var resp atlantic.ListProofsResponse
	if err := c.doRequest(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
