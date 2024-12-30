package jsonrpcws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/Azure/go-autorest/autorest"
	"github.com/gorilla/websocket"
	"github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc"
	ws "github.com/kkrt-labs/kakarot-controller/pkg/websocket"
)

// Client is a JSON-RPC client that communicates over a WebSocket connection.
type Client struct {
	client *ws.Client

	mux       sync.Mutex
	inflights map[interface{}]*operation

	wg sync.WaitGroup

	closeOnce sync.Once
	closeErr  error
	closed    chan struct{}
}

// NewClient creates a new JSON-RPC WebSocket client.
func NewClient(cfg *Config) *Client {
	var dialer ws.Dialer = ws.NewDialer(cfg.Dialer)
	dialer = ws.WithError()(dialer)
	dialer = ws.WithHeaders(http.Header{})(dialer)
	dialer = ws.WithBaseURL(cfg.Address)(dialer)

	return &Client{
		client: ws.NewClient(
			dialer,
			func(r io.Reader) (interface{}, error) { return jsonrpc.DecodeResponseMsg(r) },
		),
		inflights: make(map[interface{}]*operation),
		closed:    make(chan struct{}),
	}
}

func (c *Client) Start(ctx context.Context) error {
	err := c.client.Start(ctx)
	if err != nil {
		return err
	}

	go func() {
		c.loop()
		c.wg.Done()
	}()

	return nil
}

// Call sends a JSON-RPC request and waits for a response.
func (c *Client) Call(ctx context.Context, r *jsonrpc.Request, res interface{}) error {
	return c.call(ctx, r, res)
}

// decode decodes a JSON-RPC response message from an incoming Websocket messag
type operation struct {
	result chan *jsonrpc.ResponseMsg
}

func (c *Client) call(ctx context.Context, r *jsonrpc.Request, res interface{}) error {
	if r.ID == nil {
		return errorf(r, "missing request ID")
	}

	var err error
	r.ID, err = normalizeID(r.ID)
	if err != nil {
		return errorf(r, "%v", err)
	}

	op := &operation{
		result: make(chan *jsonrpc.ResponseMsg, 1),
	}
	defer close(op.result)

	c.mux.Lock()
	c.inflights[r.ID] = op
	c.mux.Unlock()
	defer func() {
		c.mux.Lock()
		delete(c.inflights, r.ID)
		c.mux.Unlock()
	}()

	err = c.client.SendMessage(
		ctx,
		websocket.BinaryMessage,
		func(w io.Writer) error { return json.NewEncoder(w).Encode(r) },
	)
	if err != nil {
		c.mux.Lock()
		delete(c.inflights, r.ID)
		c.mux.Unlock()
		return errorWithErrorf(err, r, "SendMessage failed")
	}

	select {
	case msg := <-op.result:
		return errorWithErrorf(msg.Unmarshal(res), r, "Failed to unmarshal response")
	case <-ctx.Done():
		return errorWithErrorf(ctx.Err(), r, "Context canceled")
	case <-c.closed:
		return errorf(r, "Client has closed")
	}
}

func errorWithErrorf(err error, r *jsonrpc.Request, message string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	msg, _ := json.Marshal(r)
	return autorest.NewErrorWithError(err, "jsonrpcws.Client", fmt.Sprintf("Call(%v)", string(msg)), nil, message, args...)
}

func errorf(r *jsonrpc.Request, message string, args ...interface{}) error {
	msg, _ := json.Marshal(r)
	return autorest.NewError("jsonrpcws.Client", fmt.Sprintf("Call(%v)", string(msg)), message, args...)
}

func (c *Client) handleIncomingMessage(msg *ws.IncomingMessage) error {
	if msg.Err() != nil {
		return msg.Err()
	}

	resp, ok := msg.Value().(*jsonrpc.ResponseMsg)
	if !ok {
		// This should never happen
		return fmt.Errorf("unexpected message value type: %T", msg.Value())
	}

	if resp.ID == nil {
		// This should never happen
		return fmt.Errorf("missing response ID")
	}

	var err error
	resp.ID, err = normalizeID(resp.ID)
	if err != nil {
		// This should never happen
		return err
	}

	c.mux.Lock()
	op, ok := c.inflights[resp.ID]
	if ok {
		// we need to return the response with the lock held
		// to ensure the channel is not closed by call() due to context cancellation
		op.result <- resp
	}
	c.mux.Unlock()

	if !ok {
		return fmt.Errorf("unknown operation ID: %v", resp.ID)
	}

	return nil
}

func (c *Client) loop() {
	for msg := range c.client.Messages() {
		_ = c.handleIncomingMessage(msg)
	}
}

func (c *Client) Errors() <-chan error {
	return c.client.Errors()
}

func (c *Client) Stop(ctx context.Context) error {
	c.closeOnce.Do(func() {
		c.closeErr = c.client.Stop(ctx)
		c.wg.Wait()
		close(c.closed)
	})

	return c.closeErr
}

// Takes an ID as received on the wire, validates it, and translates it to a
// normalized ID appropriate for keying.
func normalizeID(id interface{}) (interface{}, error) {
	switch v := id.(type) {
	case string, float64, nil:
		return v, nil
	case int64: // clients sending int64 need to normalize to float64
		return float64(v), nil
	case uint32:
		return float64(v), nil
	default:
		return nil, fmt.Errorf("invalid id type: %T (must be one of string, float64, int64, uint32)", id)
	}
}
