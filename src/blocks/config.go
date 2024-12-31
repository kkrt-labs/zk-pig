package blocks

import (
	"fmt"
	"math/big"

	jsonrpcmrgd "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/merged"
	"github.com/kkrt-labs/kakarot-controller/src/config"
)

type ChainConfig struct {
	ID  *big.Int
	RPC *jsonrpcmrgd.Config
}

// Config is the configuration for the RPCPreflight.
type Config struct {
	Chain   ChainConfig
	BaseDir string `json:"blocks-dir"` // Base directory for storing block data
}

func (cfg *Config) SetDefault() *Config {
	if cfg.BaseDir == "" {
		cfg.BaseDir = "data/blocks"
	}

	if cfg.Chain.RPC != nil {
		cfg.Chain.RPC.SetDefault()
	}

	return cfg
}

func FromGlobalConfig(gcfg *config.Config) (*Service, error) {
	cfg := &Config{
		Chain:   ChainConfig{},
		BaseDir: gcfg.DataDir,
	}

	if gcfg.Chain.ID != "" {
		cfg.Chain.ID = new(big.Int)
		if _, ok := cfg.Chain.ID.SetString(gcfg.Chain.ID, 10); !ok {
			return nil, fmt.Errorf("invalid chain id %q", gcfg.Chain.ID)
		}
	}

	if gcfg.Chain.RPC.URL != "" {
		cfg.Chain.RPC = &jsonrpcmrgd.Config{
			Addr: gcfg.Chain.RPC.URL,
		}
	}

	return New(cfg)
}
