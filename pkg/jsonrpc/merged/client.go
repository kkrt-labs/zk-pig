package jsonrpcmrgd

import (
	"fmt"

	"github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc"
	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	jsonrpcws "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/websocket"
	comurl "github.com/kkrt-labs/kakarot-controller/pkg/net/url"
)

// Config is a configuration for a JSON-RPC client
type Config struct {
	// Addr of the JSON-RPC server
	// should be a valid URL with a scheme of either http, https, ws or wss
	Addr string `json:"addr"`
	// HTTP is a configuration for HTTP client MUST be provided if addr scheme is http or https
	HTTP *jsonrpchttp.Config `json:"http"`
	// WS is a configuration for WebSocket client MUST be provided if addr scheme is ws or wss
	WS *jsonrpcws.Config `json:"ws"`
}

// SetDefault sets the default values for the configuration
func (cfg *Config) SetDefault() *Config {
	if cfg.HTTP == nil {
		cfg.HTTP = new(jsonrpchttp.Config)
	}
	cfg.HTTP.SetDefault()

	if cfg.WS == nil {
		cfg.WS = new(jsonrpcws.Config)
	}
	cfg.WS.SetDefault()

	return cfg
}

// New creates a new client capable of connecting to a JSON-RPC server either over HTTP or WebSocket
// The function returns an error if the address is invalid or the configuration is missing
func New(cfg *Config) (jsonrpc.Client, error) {
	u, err := comurl.Parse(cfg.Addr)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		if cfg.HTTP == nil {
			return nil, fmt.Errorf("HTTP configuration is required for HTTP connection")
		}

		httpc, err := jsonrpchttp.NewClient(cfg.Addr, cfg.HTTP)
		if err != nil {
			return nil, err
		}

		return httpc, nil
	}

	if u.Scheme == "ws" || u.Scheme == "wss" {
		if cfg.WS == nil {
			return nil, fmt.Errorf("WebSocket configuration is required for WebSocket connection")
		}

		wsc, err := jsonrpcws.NewClient(cfg.Addr, cfg.WS)
		if err != nil {
			return nil, err
		}

		return wsc, nil
	}

	return nil, fmt.Errorf("unsupported scheme for connection: %s", u.Scheme)
}
