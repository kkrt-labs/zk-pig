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
	storeinputs "github.com/kkrt-labs/kakarot-controller/pkg/store"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/s3"
	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
)

// Common test structures and helpers
type testCase struct {
	name            string
	contentType     storeinputs.ContentType
	contentEncoding storeinputs.ContentEncoding
	storage         string
	s3Config        *s3.Config
}

var testCases = []testCase{
	{
		name:            "JSON Plain File",
		contentType:     storeinputs.ContentTypeJSON,
		contentEncoding: storeinputs.ContentEncodingPlain,
		storage:         "file",
	},
	{
		name:            "Protobuf Plain File",
		contentType:     storeinputs.ContentTypeProtobuf,
		contentEncoding: storeinputs.ContentEncodingPlain,
		storage:         "file",
	},
	{
		name:            "JSON Gzip File",
		contentType:     storeinputs.ContentTypeJSON,
		contentEncoding: storeinputs.ContentEncodingGzip,
		storage:         "file",
	},
	{
		name:            "Protobuf Gzip File",
		contentType:     storeinputs.ContentTypeProtobuf,
		contentEncoding: storeinputs.ContentEncodingGzip,
		storage:         "file",
	},
	// TODO: Add S3 test cases
	// TODO: Figure out access key and secret key access
	// {
	// 	name:            "JSON Plain S3",
	// 	contentType:     storeinputs.ContentTypeJSON,
	// 	contentEncoding: storeinputs.ContentEncodingPlain,
	// 	storage:         "s3",
	// 	s3Config: &s3.Config{
	// 		Bucket:    "kkrt-dev-prover-inputs-s3-euw1-prover-inputs",
	// 		Region:    "eu-west-1",
	// 		AccessKey: "access-key",
	// 		SecretKey: "secret-key",
	// 		KeyPrefix: "test",
	// 	},
	// },
}

func setupProverInputsTestStore(t *testing.T, tc testCase) (store *ProverInputsStore, baseDir string) {
	baseDir = t.TempDir()
	cfg := &Config{
		MultiConfig: multi.Config{
			FileConfig: &file.Config{
				DataDir: baseDir,
			},
			S3Config: tc.s3Config,
		},
		ContentType:     tc.contentType,
		ContentEncoding: tc.contentEncoding,
	}

	store, err := New(cfg)
	assert.NoError(t, err)
	return store, baseDir
}

func createTestHeaders(tc testCase) storeinputs.Headers {
	headers := storeinputs.Headers{
		ContentType:     tc.contentType,
		ContentEncoding: tc.contentEncoding,
		KeyValue:        map[string]string{"storage": tc.storage, "key-prefix": "test"},
	}

	if tc.storage == "s3" {
		headers.KeyValue["s3-bucket"] = tc.s3Config.Bucket
		headers.KeyValue["region"] = tc.s3Config.Region
		headers.KeyValue["access-key"] = tc.s3Config.AccessKey
		headers.KeyValue["secret-key"] = tc.s3Config.SecretKey
	}

	return headers
}

func setupHeavyProverInputsTestStore(t *testing.T) (store *HeavyProverInputsStore, baseDir string) {
	cfg := &Config{
		MultiConfig: multi.Config{
			FileConfig: &file.Config{DataDir: baseDir},
		},
	}

	store, err := NewHeavyProverInputsStore(cfg)
	assert.NoError(t, err)
	return store, baseDir
}

func TestBlockStore(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			heavyProverInputsStore, _ := setupHeavyProverInputsTestStore(t)
			headers := createTestHeaders(tc)

			// Test HeavyProverInputs
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

			// Test storing and loading HeavyProverInputs
			err := heavyProverInputsStore.StoreHeavyProverInputs(context.Background(), heavyProverInputs)
			assert.NoError(t, err)

			loaded, err := heavyProverInputsStore.LoadHeavyProverInputs(context.Background(), 1, 10)
			assert.NoError(t, err)
			assert.Equal(t, heavyProverInputs.ChainConfig.ChainID, loaded.ChainConfig.ChainID)

			// Test non-existent HeavyProverInputs
			_, err = heavyProverInputsStore.LoadHeavyProverInputs(context.Background(), 1, 20)
			assert.Error(t, err)

			// Test ProverInputs
			proverInputsStore, _ := setupProverInputsTestStore(t, tc)

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

			// Test storing and loading ProverInputs
			err = proverInputsStore.StoreProverInputs(context.Background(), proverInputs, headers)
			assert.NoError(t, err)

			loadedProverInputs, err := proverInputsStore.LoadProverInputs(context.Background(), 2, 15, headers)
			assert.NoError(t, err)
			assert.Equal(t, proverInputs.ChainConfig.ChainID, loadedProverInputs.ChainConfig.ChainID)

			// Test non-existent ProverInputs
			_, err = proverInputsStore.LoadProverInputs(context.Background(), 2, 25, headers)
			assert.Error(t, err)
		})
	}
}
