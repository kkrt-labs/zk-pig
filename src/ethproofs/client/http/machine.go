package http

import (
	"context"
	"net/http"

	ethproofs "github.com/kkrt-labs/kakarot-controller/src/ethproofs/client"
)

func (c *Client) CreateMachine(ctx context.Context, req *ethproofs.CreateMachineRequest) (*ethproofs.CreateMachineResponse, error) {
	var resp ethproofs.CreateMachineResponse
	if err := c.do(ctx, http.MethodPost, "/single-machine", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
