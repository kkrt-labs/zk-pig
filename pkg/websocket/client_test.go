package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type Msg struct {
	Data string `json:"data"`
}

type handler struct {
	*testing.T
	reqs   chan *Msg
	resps  chan *Msg
	errors chan error
}

func (h handler) getNextReq(timeout time.Duration) (*Msg, error) {
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
		reqs:   make(chan *Msg),
		resps:  make(chan *Msg),
		errors: make(chan error),
	}
	return httptest.NewServer(h), h
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := (&websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			http.Error(w, reason.Error(), status)
		},
	}).Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("test-server: Upgrade: %v\n", err)
		return
	}

	go func() {
		for {
			reqMsg := new(Msg)
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

func newClient(u string) *Client {
	u = "ws" + strings.TrimPrefix(u, "http")
	dialer := &websocket.Dialer{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 30 * time.Second,
	}
	return NewClient(WithBaseURL(u)(dialer), func(r io.Reader) (interface{}, error) {
		msg := new(Msg)
		err := json.NewDecoder(r).Decode(msg)
		return msg, err
	})
}

func TestWsClientStartStop(t *testing.T) {
	s, h := newServer(t)
	defer s.Close()

	client := newClient(s.URL)
	err := client.Start(context.TODO())
	require.NoError(t, err, "Start should not error")

	err = client.Stop(context.Background())
	require.NoError(t, err, "Stop should not error")

	closeErr, err := getNextError(h.errors, 1*time.Second)
	require.NoError(t, err, "Should receive an error")

	require.IsType(t, &websocket.CloseError{}, closeErr, "unexpected error type")
	require.Equal(t, websocket.CloseNormalClosure, closeErr.(*websocket.CloseError).Code, "unexpected error code")
}

func TestWsClientSendReceiveMsg(t *testing.T) {
	s, h := newServer(t)
	defer s.Close()

	client := newClient(s.URL)
	err := client.Start(context.TODO())
	require.NoError(t, err, "Start should not error")

	t.Run("send message", func(t *testing.T) {
		client.SendMessage(context.TODO(), websocket.TextMessage, func(w io.Writer) error { return json.NewEncoder(w).Encode(&Msg{Data: "hello"}) })
		req, err := h.getNextReq(time.Second)
		require.NoError(t, err, "getNextReq should not error")
		require.Equal(t, "hello", req.Data, "unexpected request data")
	})

	t.Run("receive message", func(t *testing.T) {
		h.resps <- &Msg{Data: "world"}
		msg, err := getNextMsg(client.Messages(), time.Second)
		require.NoError(t, err, "Should receive a message")
		require.Equal(t, "world", msg.(*Msg).Data, "unexpected response data")
	})

	err = client.Stop(context.Background())
	require.NoError(t, err, "Stop should not error")
}

func getNextMsg(ch <-chan *IncomingMessage, timeout time.Duration) (interface{}, error) {
	select {
	case msg := <-ch:
		return msg.Value().(*Msg), nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for message")
	}
}

func getNextError(ch <-chan error, timeout time.Duration) (error, error) {
	select {
	case err := <-ch:
		return err, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for error")
	}
}
