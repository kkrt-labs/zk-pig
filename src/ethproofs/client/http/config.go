package http

import (
	comhttp "github.com/kkrt-labs/kakarot-controller/pkg/net/http"
)

const (
	defaultAddr = "https://staging--ethproofs.netlify.app/api/v0"
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

type Option func(*Config)

func WithAddr(addr string) Option {
	return func(c *Config) {
		c.Addr = addr
	}
}

func WithAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

func WithHTTPConfig(cfg *comhttp.ClientConfig) Option {
	return func(c *Config) {
		c.HTTPConfig = cfg
	}
}
