package ethereum

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
)

// ChainConfigs are supported chain configurations.
var chainConfigs = map[string]*params.ChainConfig{
	params.MainnetChainConfig.ChainID.String(): params.MainnetChainConfig,
	params.SepoliaChainConfig.ChainID.String(): params.SepoliaChainConfig,
	params.HoleskyChainConfig.ChainID.String(): params.HoleskyChainConfig,
}

func GetChainConfig(chainID *big.Int) (*params.ChainConfig, error) {
	cfg, ok := chainConfigs[chainID.String()]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %q", chainID.String())
	}
	return cfg, nil
}

var defaultGenesis = map[string]*core.Genesis{
	params.MainnetChainConfig.ChainID.String(): core.DefaultGenesisBlock(),
	params.SepoliaChainConfig.ChainID.String(): core.DefaultSepoliaGenesisBlock(),
	params.HoleskyChainConfig.ChainID.String(): core.DefaultHoleskyGenesisBlock(),
}

func GetDefaultGenesis(chainID *big.Int) (*core.Genesis, error) {
	genesis, ok := defaultGenesis[chainID.String()]
	if !ok {
		return nil, fmt.Errorf("unsupported chain: %q", chainID.String())
	}
	return genesis, nil
}
