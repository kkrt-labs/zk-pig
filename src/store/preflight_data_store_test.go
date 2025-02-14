package store

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/kkrt-labs/go-utils/ethereum/rpc"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	"github.com/kkrt-labs/zk-pig/src/generator"
	"github.com/stretchr/testify/assert"
)

func setupPreflightDataTestStore(t *testing.T) (store PreflightDataStore, baseDir string) {
	baseDir = t.TempDir()
	cfg := &PreflightDataStoreConfig{
		FileConfig: &filestore.Config{DataDir: baseDir},
	}

	store, err := NewPreflightDataStore(cfg)
	assert.NoError(t, err)
	return store, baseDir
}

func TestPreflightDataStore(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			preflightDataStore, _ := setupPreflightDataTestStore(t)

			// Test PreflightData
			preflightData := &generator.PreflightData{
				ChainConfig: &params.ChainConfig{
					ChainID: big.NewInt(1),
				},
				Block: &rpc.Block{
					Header: rpc.Header{
						Number: (*hexutil.Big)(hexutil.MustDecodeBig("0xa")),
					},
				},
			}

			// Test storing and loading PreflightData
			err := preflightDataStore.StorePreflightData(context.Background(), preflightData)
			assert.NoError(t, err)

			loaded, err := preflightDataStore.LoadPreflightData(context.Background(), 1, 10)
			assert.NoError(t, err)
			assert.Equal(t, preflightData.ChainConfig.ChainID, loaded.ChainConfig.ChainID)

			// Test non-existent PreflightData
			_, err = preflightDataStore.LoadPreflightData(context.Background(), 1, 20)
			assert.Error(t, err)
		})
	}
}
