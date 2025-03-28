package ethereum

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/rawdb"
	gethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/stretchr/testify/assert"
)

func TestNewChain(t *testing.T) {
	var testCases = []struct {
		desc    string
		chainID *big.Int
	}{
		{desc: "mainnet", chainID: big.NewInt(1)},
		{desc: "sepolia", chainID: big.NewInt(11155111)},
		{desc: "holesky", chainID: big.NewInt(17000)},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cfg, err := GetChainConfig(tc.chainID)
			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			trieDB := triedb.NewDatabase(rawdb.NewMemoryDatabase(), &triedb.Config{HashDB: &hashdb.Config{}})
			_, err = NewChain(cfg, gethstate.NewDatabase(trieDB, nil))
			assert.NoError(t, err)
		})
	}
}
