package generator

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

// ChainConfigs are supported chain configurations.
var ChainConfigs = map[string]*params.ChainConfig{
	params.MainnetChainConfig.ChainID.String(): params.MainnetChainConfig,
	params.SepoliaChainConfig.ChainID.String(): params.SepoliaChainConfig,
	params.HoleskyChainConfig.ChainID.String(): params.HoleskyChainConfig,
}

func getChainConfig(chainID *big.Int) (*params.ChainConfig, error) {
	cfg, ok := ChainConfigs[chainID.String()]
	if !ok {
		return nil, fmt.Errorf("unsupported chain ID: %s", chainID)
	}
	return cfg, nil
}
