package http

import (
	comhttp "github.com/kkrt-labs/kakarot-controller/pkg/net/http"
)

const (
	defaultAddr = "https://atlantic.api.herodotus.cloud"
)

type Config struct {
	Addr       string
	APIKey     string
	HTTPConfig *comhttp.ClientConfig
}

func (cfg *Config) SetDefault() *Config {
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}

	if cfg.HTTPConfig == nil {
		cfg.HTTPConfig = new(comhttp.ClientConfig)
	}

	cfg.HTTPConfig.SetDefault()

	return cfg
}
