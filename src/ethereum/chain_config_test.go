package ethereum

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainConfigAndGenesis(t *testing.T) {
	var testCases = []struct {
		desc    string
		chainID *big.Int
	}{
		{
			desc:    "mainnet",
			chainID: big.NewInt(1),
		},
		{
			desc:    "sepolia",
			chainID: big.NewInt(11155111),
		},
		{
			desc:    "holesky",
			chainID: big.NewInt(17000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cfg, err := GetChainConfig(tc.chainID)
			assert.NoError(t, err)
			assert.NotNil(t, cfg)

			genesis, err := GetDefaultGenesis(tc.chainID)
			assert.NoError(t, err)
			assert.NotNil(t, genesis)
		})
	}
}
