package jsonrpchttp

import (
	"context"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc"
	httptestutils "github.com/kkrt-labs/kakarot-controller/pkg/net/http/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestClientImplementsInterface(t *testing.T) {
	assert.Implementsf(t, (*jsonrpc.Client)(nil), new(Client), "Client should implement jsonrpc.Client")
}

func TestCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCli := httptestutils.NewMockSender(ctrl)
	c := NewClientFromClient(mockCli)

	t.Run("StatusOKAndValidResult", func(t *testing.T) { testCallStatusOKAndValidResult(t, c, mockCli) })
	t.Run("StatusOKAndError", func(t *testing.T) { testCallStatusOKAndError(t, c, mockCli) })
	t.Run("Status400", func(t *testing.T) { testCallStatus400(t, c, mockCli) })
}

func testCallStatusOKAndValidResult(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"2.0","method":"concat","params":["a","b","c"],"id":0}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","result":"abc","id":0}`))

	mockCli.EXPECT().Gock(req)

	var res string
	err := c.Call(
		context.Background(),
		&jsonrpc.Request{
			Version: "2.0",
			Method:  "concat",
			Params:  []string{"a", "b", "c"},
			ID:      0,
		},
		&res,
	)

	require.NoError(t, err)
	assert.Equal(
		t,
		"abc",
		res,
	)
}

func testCallStatusOKAndError(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"2.0","method":"concat","params":["a","b","c"],"id":0}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","error":{"code":-32000,"message":"invalid test method"},"id":0}`))

	mockCli.EXPECT().Gock(req)

	var res string
	err := c.Call(
		context.Background(),
		&jsonrpc.Request{
			Version: "2.0",
			Method:  "concat",
			Params:  []string{"a", "b", "c"},
			ID:      0,
		},
		&res,
	)

	require.Error(t, err)
	require.IsType(t, autorest.DetailedError{}, err)
	assert.Equal(
		t,
		&jsonrpc.ErrorMsg{
			Code:    -32000,
			Message: "invalid test method",
		},
		err.(autorest.DetailedError).Original,
	)
}

func testCallStatus400(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		Reply(400)

	mockCli.EXPECT().Gock(req)

	var res string
	err := c.Call(
		context.Background(),
		&jsonrpc.Request{
			Version: "2.0",
			Method:  "concat",
			Params:  []string{"a", "b", "c"},
			ID:      0,
		},
		&res,
	)

	require.Error(t, err)
}
