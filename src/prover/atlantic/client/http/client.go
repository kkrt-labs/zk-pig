package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
	comhttp "github.com/kkrt-labs/go-utils/net/http"
)

type Client struct {
	client autorest.Sender
	cfg    *Config
}

func NewClient(cfg *Config) (*Client, error) {
	cfg.SetDefault()

	httpc, err := comhttp.NewClient(cfg.HTTPConfig)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &Client{
		client: autorest.Client{
			Sender:           httpc,
			RequestInspector: comhttp.WithBaseURL(baseURL),
		},
		cfg: cfg,
	}, nil
}

func (c *Client) prepareRequest(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Request, error) {
	preparers := []autorest.PrepareDecorator{
		autorest.WithMethod(method),
		autorest.WithPath(path),
	}

	if contentType != "" {
		preparers = append(preparers, autorest.AsContentType(contentType))
	}

	if body != nil {
		preparers = append(preparers, autorest.WithFile(io.NopCloser(body)))
	}

	if c.cfg.APIKey != "" {
		preparers = append(preparers, autorest.WithQueryParameters(map[string]interface{}{
			"apiKey": c.cfg.APIKey,
		}))
	}

	req, err := autorest.CreatePreparer(preparers...).Prepare(newRequest(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	return req, nil
}

func newRequest(ctx context.Context) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, "", "", http.NoBody)
	return req
}

func (c *Client) doRequest(req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}

	err = autorest.Respond(
		resp,
		autorest.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(v),
		autorest.ByClosing(),
	)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	return nil
}
