package store

import (
	"bytes"
	"context"
	"io"
	"math/big"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/kkrt-labs/go-utils/store"
	mockstore "github.com/kkrt-labs/go-utils/store/mock"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProverInputStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockstore.NewMockStore(ctrl)

	testCases := []struct {
		desc        string
		contentType store.ContentType
		chainID     int64
		blockNumber int64
		expectedKey string
	}{
		{
			desc:        "JSON Plain File",
			contentType: store.ContentTypeJSON,
			chainID:     2,
			blockNumber: 15,
			expectedKey: "/2/inputs/15.json",
		},
		{
			desc:        "Protobuf Plain File",
			contentType: store.ContentTypeProtobuf,
			chainID:     2,
			blockNumber: 15,
			expectedKey: "/2/inputs/15.protobuf",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			// Test ProverInput
			inputStore := NewProverInputStore(mockStore, tt.contentType)

			in := &input.ProverInput{
				ChainConfig: &params.ChainConfig{
					ChainID: big.NewInt(tt.chainID),
				},
				Blocks: []*input.Block{
					{
						Header: &gethtypes.Header{
							Number:          big.NewInt(tt.blockNumber),
							Difficulty:      big.NewInt(15),
							BaseFee:         big.NewInt(15),
							WithdrawalsHash: &gethcommon.Hash{0x1},
						},
					},
				},
			}

			// Test storing and loading ProverInput
			var dataCache []byte
			ctx := context.TODO()
			mockStore.EXPECT().Store(ctx, tt.expectedKey, gomock.Any(), &store.Headers{
				ContentType: tt.contentType,
			}).DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ *store.Headers) error {
				dataCache, _ = io.ReadAll(reader)
				return nil
			})

			err := inputStore.StoreProverInput(ctx, in)
			assert.NoError(t, err)

			mockStore.EXPECT().Load(ctx, tt.expectedKey).Return(io.NopCloser(bytes.NewReader(dataCache)), nil, nil)
			loadedProverInput, err := inputStore.LoadProverInput(ctx, 2, 15)
			assert.NoError(t, err)
			assert.Equal(t, in.ChainConfig.ChainID, loadedProverInput.ChainConfig.ChainID)
			assert.Equal(t, in.Blocks[0].Header.Number, loadedProverInput.Blocks[0].Header.Number)
		})
	}
}

func TestNoOpProverInputStore(t *testing.T) {
	noOpStore := NewNoOpProverInputStore()
	// Should implement interface
	assert.Implements(t, (*ProverInputStore)(nil), noOpStore)
	assert.NoError(t, noOpStore.StoreProverInput(context.TODO(), &input.ProverInput{}))

	loaded, err := noOpStore.LoadProverInput(context.TODO(), 1, 1)
	assert.Nil(t, loaded)
	assert.NoError(t, err)
}
