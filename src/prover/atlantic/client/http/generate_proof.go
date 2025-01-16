package http

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"

	atlantic "github.com/kkrt-labs/kakarot-controller/src/prover/atlantic/client"
)

func (c *Client) GenerateProof(ctx context.Context, req *atlantic.GenerateProofRequest) (*atlantic.GenerateProofResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add pie file
	part, err := writer.CreateFormFile("pieFile", "proof.pie")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, bytes.NewReader(req.PieFile)); err != nil {
		return nil, err
	}

	// Add other form fields
	if err := writer.WriteField("layout", string(req.Layout)); err != nil {
		return nil, err
	}
	if err := writer.WriteField("prover", string(req.Prover)); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	httpReq, err := c.prepareRequest(ctx, http.MethodPost, "/v1/proof-generation", body, writer.FormDataContentType())
	if err != nil {
		return nil, err
	}

	var resp atlantic.GenerateProofResponse
	if err := c.doRequest(httpReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
