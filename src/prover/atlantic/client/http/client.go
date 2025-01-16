package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	comhttp "github.com/kkrt-labs/kakarot-controller/pkg/net/http"
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

	return &Client{
		client: autorest.Client{
			Sender:           httpc,
			RequestInspector: comhttp.WithBaseURL(cfg.Addr),
		},
		cfg: cfg,
	}, nil
}

func (c *Client) prepareRequest(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.cfg.APIKey != "" {
		req.Header.Set("apiKey", c.cfg.APIKey)
	}

	return req, nil
}

func (c *Client) doRequest(req *http.Request, v interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
