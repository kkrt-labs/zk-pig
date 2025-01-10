package http

import (
	"context"
	"net/http"

	"github.com/kkrt-labs/kakarot-controller/src/ethproofs"
)

func (c *Client) CreateMachine(ctx context.Context, req *ethproofs.CreateMachineRequest) (*ethproofs.CreateMachineResponse, error) {
	httpReq, err := c.newRequest(ctx, http.MethodPost, "/single-machine", req)
	if err != nil {
		return nil, err
	}

	var resp ethproofs.CreateMachineResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
