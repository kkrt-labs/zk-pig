package http

import (
	"context"
	"net/http"

	"github.com/kkrt-labs/kakarot-controller/src/ethproofs"
)

func (c *Client) ListAWSPricing(ctx context.Context) ([]ethproofs.AWSInstance, error) {
	httpReq, err := c.newRequest(ctx, http.MethodGet, "/aws-pricing-list", nil)
	if err != nil {
		return nil, err
	}

	var resp []ethproofs.AWSInstance
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
