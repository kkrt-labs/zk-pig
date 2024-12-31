package jsonrpcws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientImplementsInterface(t *testing.T) {
	assert.Implementsf(t, (*jsonrpc.Client)(nil), new(Client), "Client should implement jsonrpc.Client")
}

type handler struct {
	*testing.T
	reqs   chan *jsonrpc.RequestMsg
	resps  chan *jsonrpc.ResponseMsg
	errors chan error
}

func (h handler) getNextReq(timeout time.Duration) (*jsonrpc.RequestMsg, error) {
	select {
	case req := <-h.reqs:
		return req, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for request")
	}
}

func newServer(t *testing.T) (*httptest.Server, handler) {
	h := handler{
		T:      t,
		reqs:   make(chan *jsonrpc.RequestMsg),
		resps:  make(chan *jsonrpc.ResponseMsg),
		errors: make(chan error),
	}
	return httptest.NewServer(h), h
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := (&websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		Error: func(w http.ResponseWriter, _ *http.Request, status int, reason error) {
			http.Error(w, reason.Error(), status)
		},
	}).Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("test-server: Upgrade: %v\n", err)
		return
	}

	go func() {
		for {
			reqMsg := new(jsonrpc.RequestMsg)
			e := conn.ReadJSON(reqMsg)
			if e != nil {
				fmt.Printf("test-server: ReadJSON: resp=%v err=%v\n", *reqMsg, e)
				h.errors <- e
				return
			}
			h.reqs <- reqMsg
		}
	}()

	go func() {
		defer conn.Close()
		for resp := range h.resps {
			err = conn.WriteJSON(resp)
			if err != nil {
				fmt.Printf("test-server: WriteJSON: %v\n", err)
				return
			}
		}
	}()
}

func newClient(t *testing.T, u string) *Client {
	u = "ws" + strings.TrimPrefix(u, "http")
	c, err := NewClient(u, (&Config{}).SetDefault())
	require.NoError(t, err, "NewClient should not error")
	return c
}

type testCtx struct {
	t *testing.T

	s *httptest.Server
	c *Client
	h handler
}

func newTestCtx(t *testing.T) *testCtx {
	server, h := newServer(t)
	client := newClient(t, server.URL)

	err := client.Start(context.TODO())
	require.NoError(t, err, "Start should not error")

	return &testCtx{t: t, s: server, c: client, h: h}
}

func TestWsClientSendReceiveMsg(t *testing.T) {
	t.Run("StartStop", func(t *testing.T) {
		tt := newTestCtx(t)
		defer tt.s.Close()

		err := tt.c.Stop(context.Background())
		require.NoError(t, err, "Stop should not error")

		closeErr, err := getNextError(tt.h.errors, 1*time.Second)
		require.NoError(t, err, "Should receive an error")

		require.IsType(t, &websocket.CloseError{}, closeErr, "unexpected error type")
		require.Equal(t, websocket.CloseNormalClosure, closeErr.(*websocket.CloseError).Code, "unexpected error code")
	})

	t.Run("Call", func(t *testing.T) {
		tt := newTestCtx(t)
		defer tt.s.Close()

		req := &jsonrpc.Request{ID: 1, Method: "test", Params: []string{"hello"}}
		res := new(string)

		done := make(chan struct{})
		var callErr error
		go func() {
			callErr = tt.c.Call(context.TODO(), req, res)
			close(done)
		}()

		// Check the server received the request
		srvReq, err := tt.h.getNextReq(time.Second)
		require.NoError(t, err, "getNextReq should not error")
		gotB, _ := json.Marshal(srvReq)
		expectedB, _ := json.Marshal(req)
		assert.JSONEq(t, string(expectedB), string(gotB))

		// Server sends a response with same ID
		tt.h.resps <- &jsonrpc.ResponseMsg{ID: 1, Result: json.RawMessage(`"world"`)}
		err = waitDone(done, time.Second)
		require.NoError(t, err, "Call should have ended")
		assert.NoError(t, callErr, "Call should not error")
		require.NotNil(t, res, "response should not be nil")
		assert.Equal(t, "world", *res, "unexpected response data")

		err = tt.c.Stop(context.Background())
		require.NoError(t, err, "Stop should not error")
	})

	t.Run("Call#WithTimeout", func(t *testing.T) {
		tt := newTestCtx(t)
		defer tt.s.Close()

		res := new(string)
		done := make(chan struct{})
		var callErr error
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			callErr = tt.c.Call(ctx, &jsonrpc.Request{ID: 1, Method: "test", Params: []string{"hello"}}, res)
			close(done)
		}()

		// Check the server received the request
		_, err := tt.h.getNextReq(time.Second)
		require.NoError(t, err, "getNextReq should not error")

		// Sleeps before sending response so the timeout triggers
		time.Sleep(350 * time.Millisecond)
		tt.h.resps <- &jsonrpc.ResponseMsg{ID: 1, Result: json.RawMessage(`"world"`)}
		err = waitDone(done, time.Second)
		require.NoError(t, err, "Call should have ended")
		assert.Error(t, callErr, "Call should have error")

		err = tt.c.Stop(context.Background())
		require.NoError(t, err, "Stop should not error")
	})

	t.Run("Call#ClientClose", func(t *testing.T) {
		tt := newTestCtx(t)
		defer tt.s.Close()

		res := new(string)
		done := make(chan struct{})
		var callErr error
		go func() {
			callErr = tt.c.Call(context.TODO(), &jsonrpc.Request{ID: 1, Method: "test", Params: []string{"hello"}}, res)
			close(done)
		}()

		// Check the server received the request
		_, err := tt.h.getNextReq(time.Second)
		require.NoError(t, err, "getNextReq should not error")

		// Close the client before sending response
		err = tt.c.Stop(context.Background())
		require.NoError(t, err, "Stop should not error")

		// Inspect the call results
		err = waitDone(done, time.Second)
		require.NoError(t, err, "Call shouldhave ended")
		assert.Error(t, callErr, "response should be nil")
	})
}

func getNextError(ch <-chan error, timeout time.Duration) (gotError, err error) {
	select {
	case gotError := <-ch:
		return gotError, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for error")
	}
}

func waitDone(done chan struct{}, timeout time.Duration) error {
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for done")
	}
}
