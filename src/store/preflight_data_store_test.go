package store

import (
	"bytes"
	"context"
	"io"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/kkrt-labs/go-utils/ethereum/rpc"
	store "github.com/kkrt-labs/go-utils/store"
	mockstore "github.com/kkrt-labs/go-utils/store/mock"
	"github.com/kkrt-labs/zk-pig/src/steps"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPreflightDataStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockstore.NewMockStore(ctrl)
	preflightDataStore, err := NewPreflightDataStore(mockStore)
	assert.NoError(t, err)

	// Test PreflightData
	preflightData := &steps.PreflightData{
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
	var dataCache []byte
	ctx := context.TODO()
	mockStore.EXPECT().Store(ctx, "/1/preflight/10.json", gomock.Any(), &store.Headers{
		ContentType:     store.ContentTypeJSON,
		ContentEncoding: store.ContentEncodingPlain,
	}).DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ *store.Headers) error {
		dataCache, _ = io.ReadAll(reader)
		return nil
	})
	err = preflightDataStore.StorePreflightData(ctx, preflightData)
	assert.NoError(t, err)

	mockStore.EXPECT().Load(ctx, "/1/preflight/10.json").Return(io.NopCloser(bytes.NewReader(dataCache)), nil, nil)
	loaded, err := preflightDataStore.LoadPreflightData(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, preflightData.ChainConfig.ChainID, loaded.ChainConfig.ChainID)
	assert.Equal(t, preflightData.Block.Header.Number, loaded.Block.Header.Number)
}

func TestNoOpPreflightDataStore(t *testing.T) {
	noOpStore := NewNoOpPreflightDataStore()
	// Should implement interface
	assert.Implements(t, (*PreflightDataStore)(nil), noOpStore)
	assert.NoError(t, noOpStore.StorePreflightData(context.TODO(), &steps.PreflightData{}))

	loaded, err := noOpStore.LoadPreflightData(context.TODO(), 1, 1)
	assert.Nil(t, loaded)
	assert.NoError(t, err)
}
