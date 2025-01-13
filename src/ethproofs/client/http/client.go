package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	comhttp "github.com/kkrt-labs/kakarot-controller/pkg/net/http"
)

type Client struct {
	client autorest.Sender
	apiKey string
}

func NewClientFromClient(s autorest.Sender, apiKey string) *Client {
	return &Client{
		client: s,
		apiKey: apiKey,
	}
}

func NewClient(cfg *Config) (*Client, error) {
	cfg.SetDefault()

	httpc, err := comhttp.NewClient(cfg.HTTPConfig)
	if err != nil {
		return nil, err
	}

	return NewClientFromClient(
		autorest.Client{
			Sender:           httpc,
			RequestInspector: comhttp.WithBaseURL(cfg.Addr),
		},
		cfg.APIKey,
	), nil
}

func (c *Client) do(ctx context.Context, method, path string, body, res interface{}) error {
	req, err := c.prepareRequest(ctx, method, path, body)
	if err != nil {
		return autorest.NewErrorWithError(err, "ethproofs.Client", "do", nil, "PrepareRequest")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return autorest.NewErrorWithError(err, "ethproofs.Client", "do", resp, "Do")
	}

	err = c.inspectResponse(resp, res)
	if err != nil {
		return autorest.NewErrorWithError(err, "ethproofs.Client", "do", resp, "Inspect Response")
	}

	return nil
}

func (c *Client) prepareRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	preparers := []autorest.PrepareDecorator{
		autorest.AsContentType("application/json"),
		autorest.WithPath(path),
		autorest.WithMethod(method),
		autorest.WithHeader("Authorization", fmt.Sprintf("Bearer %s", c.apiKey)),
	}

	if body != nil {
		preparers = append(preparers, autorest.WithJSON(body))
	}

	return autorest.CreatePreparer(preparers...).Prepare(newRequest(ctx))
}

func newRequest(ctx context.Context) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, "", "", http.NoBody)
	return req
}

func (c *Client) inspectResponse(resp *http.Response, res interface{}) error {
	err := autorest.Respond(
		resp,
		autorest.WithErrorUnlessOK(),
		autorest.ByUnmarshallingJSON(res),
		autorest.ByClosing(),
	)
	if err != nil {
		return err
	}

	return nil
}
