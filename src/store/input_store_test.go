package store

import (
	"context"
	"math/big"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	storeinputs "github.com/kkrt-labs/go-utils/store"
	compressstore "github.com/kkrt-labs/go-utils/store/compress"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	multistore "github.com/kkrt-labs/go-utils/store/multi"
	s3store "github.com/kkrt-labs/go-utils/store/s3"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/stretchr/testify/assert"
)

// Common test structures and helpers
type testCase struct {
	name            string
	contentType     storeinputs.ContentType
	contentEncoding storeinputs.ContentEncoding
	storage         string
	s3Config        *s3store.Config
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
	// 		Bucket:    "kkrt-dev-prover-input-s3-euw1-prover-input",
	// 		Region:    "eu-west-1",
	// 		AccessKey: "access-key",
	// 		SecretKey: "secret-key",
	// 		BucketKeyPrefix: "test",
	// 	},
	// },
}

func setupProverInputTestStore(t *testing.T, tc testCase) (store ProverInputStore, baseDir string) {
	baseDir = t.TempDir()
	cfg := &ProverInputStoreConfig{
		StoreConfig: multistore.Config{
			FileConfig: &filestore.Config{
				DataDir: baseDir,
			},
			S3Config: tc.s3Config,
		},
	}
	compressStore, err := compressstore.New(compressstore.Config{
		MultiStoreConfig: cfg.StoreConfig,
		ContentEncoding:  tc.contentEncoding,
	})
	store = NewFromStore(compressStore, tc.contentType)

	assert.NoError(t, err)
	return store, baseDir
}

func TestProverInputStore(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test ProverInput
			ProverInputStore, _ := setupProverInputTestStore(t, tc)

			ProverInput := &input.ProverInput{
				ChainConfig: &params.ChainConfig{
					ChainID: big.NewInt(2),
				},
				Blocks: []*input.Block{
					{
						Header: &gethtypes.Header{
							Number:          big.NewInt(15),
							Difficulty:      big.NewInt(15),
							BaseFee:         big.NewInt(15),
							WithdrawalsHash: &gethcommon.Hash{0x1},
						},
					},
				},
			}

			// Test storing and loading ProverInput
			err := ProverInputStore.StoreProverInput(context.Background(), ProverInput)
			assert.NoError(t, err)

			loadedProverInput, err := ProverInputStore.LoadProverInput(context.Background(), 2, 15)
			assert.NoError(t, err)
			assert.Equal(t, ProverInput.ChainConfig.ChainID, loadedProverInput.ChainConfig.ChainID)

			// Test non-existent ProverInput
			_, err = ProverInputStore.LoadProverInput(context.Background(), 2, 25)
			assert.Error(t, err)
		})
	}
}
