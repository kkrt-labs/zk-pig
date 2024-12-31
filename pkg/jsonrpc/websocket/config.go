package jsonrpcws

import (
	ws "github.com/kkrt-labs/kakarot-controller/pkg/websocket"
)

type Config struct {
	Dialer *ws.DialerConfig
}

func (cfg *Config) SetDefault() *Config {
	if cfg.Dialer == nil {
		cfg.Dialer = new(ws.DialerConfig)
	}

	cfg.Dialer.SetDefault()

	return cfg
}
