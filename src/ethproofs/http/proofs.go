package http

import (
	"context"
	"net/http"

	"github.com/kkrt-labs/kakarot-controller/src/ethproofs"
)

func (c *Client) QueueProof(ctx context.Context, req *ethproofs.QueueProofRequest) (*ethproofs.ProofResponse, error) {
	httpReq, err := c.newRequest(ctx, http.MethodPost, "/proofs/queued", req)
	if err != nil {
		return nil, err
	}

	var resp ethproofs.ProofResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) StartProving(ctx context.Context, req *ethproofs.StartProvingRequest) (*ethproofs.ProofResponse, error) {
	httpReq, err := c.newRequest(ctx, http.MethodPost, "/proofs/proving", req)
	if err != nil {
		return nil, err
	}

	var resp ethproofs.ProofResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) SubmitProof(ctx context.Context, req *ethproofs.SubmitProofRequest) (*ethproofs.ProofResponse, error) {
	httpReq, err := c.newRequest(ctx, http.MethodPost, "/proofs/proved", req)
	if err != nil {
		return nil, err
	}

	var resp ethproofs.ProofResponse
	if err := c.do(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
