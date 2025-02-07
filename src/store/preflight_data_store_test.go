package store

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/kkrt-labs/go-utils/ethereum/rpc"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/stretchr/testify/assert"
)

func setupHeavyProverInputTestStore(t *testing.T) (store HeavyProverInputStore, baseDir string) {
	baseDir = t.TempDir()
	cfg := &HeavyProverInputStoreConfig{
		FileConfig: &filestore.Config{DataDir: baseDir},
	}

	store, err := NewHeavyProverInputStore(cfg)
	assert.NoError(t, err)
	return store, baseDir
}

func TestHeavyProverInputStore(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			heavyProverInputtore, _ := setupHeavyProverInputTestStore(t)

			// Test HeavyProverInput
			heavyProverInput := &input.HeavyProverInput{
				ChainConfig: &params.ChainConfig{
					ChainID: big.NewInt(1),
				},
				Block: &rpc.Block{
					Header: rpc.Header{
						Number: (*hexutil.Big)(hexutil.MustDecodeBig("0xa")),
					},
				},
			}

			// Test storing and loading HeavyProverInput
			err := heavyProverInputtore.StoreHeavyProverInput(context.Background(), heavyProverInput)
			assert.NoError(t, err)

			loaded, err := heavyProverInputtore.LoadHeavyProverInput(context.Background(), 1, 10)
			assert.NoError(t, err)
			assert.Equal(t, heavyProverInput.ChainConfig.ChainID, loaded.ChainConfig.ChainID)

			// Test non-existent HeavyProverInput
			_, err = heavyProverInputtore.LoadHeavyProverInput(context.Background(), 1, 20)
			assert.Error(t, err)
		})
	}
}
