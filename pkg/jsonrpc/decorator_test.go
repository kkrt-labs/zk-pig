package jsonrpc_test

import (
	"context"
	"testing"

	"github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc"
	jsonrpcmock "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestWithVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCli := jsonrpcmock.NewMockClient(ctrl)
	c := jsonrpc.WithVersion("2.0")(mockCli)

	mockCli.EXPECT().Call(
		gomock.Any(),
		jsonrpcmock.HasVersion("2.0"),
		gomock.Any())
	err := c.Call(context.Background(), &jsonrpc.Request{}, nil)
	require.NoError(t, err)
}

func TestWithIncrementalID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCli := jsonrpcmock.NewMockClient(ctrl)
	c := jsonrpc.WithIncrementalID()(mockCli)

	mockCli.EXPECT().Call(
		gomock.Any(),
		jsonrpcmock.HasID(uint32(0)),
		gomock.Any())
	err := c.Call(context.Background(), &jsonrpc.Request{}, nil)
	require.NoError(t, err)

	mockCli.EXPECT().Call(
		gomock.Any(),
		jsonrpcmock.HasID(uint32(1)),
		gomock.Any())
	err = c.Call(context.Background(), &jsonrpc.Request{}, nil)
	require.NoError(t, err)
}
