package jsonrpcws

import (
	ws "github.com/kkrt-labs/kakarot-controller/pkg/websocket"
)

type Config struct {
	Client *ws.ClientConfig // WebSocket client configuration
}

func (cfg *Config) SetDefault() *Config {
	if cfg.Client == nil {
		cfg.Client = new(ws.ClientConfig)
	}
	cfg.Client.SetDefault()

	return cfg
}
