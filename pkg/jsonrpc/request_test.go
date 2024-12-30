package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalRequest(t *testing.T) {
	t.Run("WithAllFields", testMarshalRequestWithAllFields)
	t.Run("WithEmptyFields", testMarshalRequestWithEmptyFields)
	t.Run("WithInterfaceSliceParams", testMarshalRequestWithInterfaceSliceParams)
}

func testMarshalRequest(t *testing.T, expected []byte, req *Request) {
	b, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Equal(t, expected, b)
}

func testMarshalRequestWithAllFields(t *testing.T) {
	req := &Request{
		Version: "2.0",
		Method:  "test-method",
		Params:  "test-params",
		ID:      0,
	}
	expected := []byte(`{"jsonrpc":"2.0","method":"test-method","params":"test-params","id":0}`)
	testMarshalRequest(t, expected, req)
}

func testMarshalRequestWithEmptyFields(t *testing.T) {
	req := &Request{}
	expected := []byte(`{"jsonrpc":"","method":""}`)
	testMarshalRequest(t, expected, req)
}

func testMarshalRequestWithInterfaceSliceParams(t *testing.T) {
	req := &Request{
		Params: []interface{}{"test-param", 4},
	}
	expected := []byte(`{"jsonrpc":"","method":"","params":["test-param",4]}`)
	testMarshalRequest(t, expected, req)
}
