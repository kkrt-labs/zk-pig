package http

import (
	"context"
	"net/http"

	ethproofs "github.com/kkrt-labs/kakarot-controller/src/ethproofs/client"
)

func (c *Client) QueueProof(ctx context.Context, req *ethproofs.QueueProofRequest) (*ethproofs.ProofResponse, error) {
	var resp ethproofs.ProofResponse
	if err := c.do(ctx, http.MethodPost, "/proofs/queued", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) StartProving(ctx context.Context, req *ethproofs.StartProvingRequest) (*ethproofs.ProofResponse, error) {
	var resp ethproofs.ProofResponse
	if err := c.do(ctx, http.MethodPost, "/proofs/proving", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) SubmitProof(ctx context.Context, req *ethproofs.SubmitProofRequest) (*ethproofs.ProofResponse, error) {
	var resp ethproofs.ProofResponse
	if err := c.do(ctx, http.MethodPost, "/proofs/proved", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
