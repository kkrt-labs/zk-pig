package store

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	"github.com/kkrt-labs/go-utils/store"
	mockstore "github.com/kkrt-labs/go-utils/store/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBlockStore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mockstore.NewMockStore(ctrl)
	blockStore := NewBlockStore(mockStore)

	// Test BlockStore
	block := &ethrpc.Block{
		Header: ethrpc.Header{
			Number: (*hexutil.Big)(hexutil.MustDecodeBig("0xa")),
		},
	}

	// Test storing and loading Block
	var dataCache []byte
	ctx := context.TODO()
	mockStore.EXPECT().Store(
		ctx,
		"/1/blocks/10.json",
		gomock.Any(),
		&store.Headers{
			ContentType:     store.ContentTypeJSON,
			ContentEncoding: store.ContentEncodingPlain,
			KeyValue: map[string]string{
				"chain.id":     "1",
				"block.number": "10",
			},
		}).DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ *store.Headers) error {
		dataCache, _ = io.ReadAll(reader)
		return nil
	}).AnyTimes()
	err := blockStore.StoreBlock(ctx, 1, block)
	assert.NoError(t, err)

	mockStore.EXPECT().Load(ctx, "/1/blocks/10.json").Return(io.NopCloser(bytes.NewReader(dataCache)), nil, nil)
	loaded, err := blockStore.LoadBlock(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, block.Number, loaded.Number)
}

func TestNoOpBlockStore(t *testing.T) {
	noOpStore := NewNoOpBlockStore()
	assert.Implements(t, (*BlockStore)(nil), noOpStore)
	assert.NoError(t, noOpStore.StoreBlock(context.TODO(), 1, &ethrpc.Block{}))

	loaded, err := noOpStore.LoadBlock(context.TODO(), 1, 1)
	assert.Nil(t, loaded)
	assert.NoError(t, err)
}
