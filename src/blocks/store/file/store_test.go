package fileblockstore

// Implement test cases for the FileBlockStore struct.

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc"
	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
	filestore "github.com/kkrt-labs/kakarot-controller/src/blocks/store"
)

func TestFileBlockStoreJSON(t *testing.T) {
	baseDir := t.TempDir()
	store := New(baseDir)

	// Test storing and loading preflight data
	HeavyProverInputs := &blockinputs.HeavyProverInputs{
		ChainConfig: &params.ChainConfig{
			ChainID: big.NewInt(1),
		},
		Block: &rpc.Block{
			Header: rpc.Header{
				Number: (*hexutil.Big)(hexutil.MustDecodeBig("0xa")),
			},
		},
	}
	err := store.StoreHeavyProverInputs(context.Background(), HeavyProverInputs)
	assert.NoError(t, err)

	_, err = store.LoadHeavyProverInputs(context.Background(), 1, 10)
	assert.NoError(t, err)

	// Test loading non-existent preflight data
	_, err = store.LoadHeavyProverInputs(context.Background(), 1, 20)
	assert.Error(t, err)

	// Test storing and loading prover inputs
	proverInputs := &blockinputs.ProverInputs{
		ChainConfig: &params.ChainConfig{
			ChainID: big.NewInt(2),
		},
		Block: &rpc.Block{
			Header: rpc.Header{
				Number: (*hexutil.Big)(hexutil.MustDecodeBig("0xf")),
			},
		},
	}
	err = store.StoreProverInputs(context.Background(), proverInputs, filestore.JSONFormat, "gzip")
	assert.NoError(t, err)

	_, err = store.LoadProverInputs(context.Background(), 2, 15, filestore.JSONFormat, "gzip")
	assert.NoError(t, err)

	// Test loading non-existent prover inputs
	_, err = store.LoadProverInputs(context.Background(), 2, 25, filestore.JSONFormat, "gzip")
	assert.Error(t, err)
}

func TestFileBlockStoreProtobuf(t *testing.T) {
	baseDir := t.TempDir()
	store := New(baseDir)

	// Test storing and loading preflight data
	heavyProverInputs := &blockinputs.HeavyProverInputs{
		ChainConfig: &params.ChainConfig{
			ChainID: big.NewInt(1),
		},
		Block: &rpc.Block{
			Header: rpc.Header{
				Number: (*hexutil.Big)(hexutil.MustDecodeBig("0xa")),
			},
		},
	}
	err := store.StoreHeavyProverInputs(context.Background(), heavyProverInputs)
	assert.NoError(t, err)

	_, err = store.LoadHeavyProverInputs(context.Background(), 1, 10)
	assert.NoError(t, err)

	// Test loading non-existent preflight data
	_, err = store.LoadHeavyProverInputs(context.Background(), 1, 20)
	assert.Error(t, err)

	// Test storing and loading prover inputs
	proverInputs := &blockinputs.ProverInputs{
		ChainConfig: &params.ChainConfig{
			ChainID: big.NewInt(2),
		},
		Block: &rpc.Block{
			Header: rpc.Header{
				Number:          (*hexutil.Big)(hexutil.MustDecodeBig("0xf")),
				Difficulty:      (*hexutil.Big)(hexutil.MustDecodeBig("0xf")),
				BaseFee:         (*hexutil.Big)(hexutil.MustDecodeBig("0xf")),
				WithdrawalsRoot: &gethcommon.Hash{0x1},
			},
		},
	}
	err = store.StoreProverInputs(context.Background(), proverInputs, filestore.ProtobufFormat, "")
	assert.NoError(t, err)

	_, err = store.LoadProverInputs(context.Background(), 2, 15, filestore.ProtobufFormat, "")
	assert.NoError(t, err)

	// Test loading non-existent prover inputs
	_, err = store.LoadProverInputs(context.Background(), 2, 25, filestore.ProtobufFormat, "")
	assert.Error(t, err)
}
