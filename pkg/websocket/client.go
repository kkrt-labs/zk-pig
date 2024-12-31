package websocket

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	comurl "github.com/kkrt-labs/kakarot-controller/pkg/net/url"
)

// ClinetConfig is the configuration for the websocket client
type ClientConfig struct {
	Dialer *DialerConfig

	// PingInterval is the interval at which the client sends ping messages to the server
	PingInterval        time.Duration
	PongTimeout         time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	WriteControlTimeout time.Duration
}

func (cfg *ClientConfig) SetDefault() *ClientConfig {
	if cfg.Dialer == nil {
		cfg.Dialer = new(DialerConfig)
	}
	cfg.Dialer.SetDefault()

	if cfg.PingInterval == 0 {
		cfg.PingInterval = 30 * time.Second
	}
	if cfg.PongTimeout == 0 {
		cfg.PongTimeout = 30 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	if cfg.WriteControlTimeout == 0 {
		cfg.WriteControlTimeout = 5 * time.Second
	}

	return cfg
}

// Client is a client to a websocket server
// It is responsible for
// - sending and receiving messages to/from the server
// - handling ping/pong heartbeat to keep the connection alive
type Client struct {
	dialer Dialer

	cfg *ClientConfig

	conn *websocket.Conn

	wg sync.WaitGroup

	stopOnce     sync.Once        // close stopped channel once
	stopped      chan interface{} // closed on Stop
	closeMsgCode int              // close message code

	closeOnce sync.Once // close conn once
	closeErr  error

	// Ping/Pong
	pingReset     chan struct{}
	outgoingPongs chan struct{}
	outgoingPings chan struct{}

	// incoming messages
	incomingMessages chan *IncomingMessage
	incomingErr      error
	decode           func(io.Reader) (interface{}, error)

	// outgoing messages
	outgoingMessages chan *outgoingMessage
	notifyErrors     chan error
}

// NewClient creates a new websocket client
//   - wsConn is the underlying websocket connection
//   - decode is the function that is used to decode incoming messages
func NewClient(addr string, cfg *ClientConfig, decode func(io.Reader) (interface{}, error)) (*Client, error) {
	u, err := comurl.Parse(addr)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "ws" && u.Scheme != "wss" {
		return nil, fmt.Errorf("unsupported scheme for websocket connection: %s", u.Scheme)
	}

	return &Client{
		dialer:           WithBaseURL(u)(NewDialer(cfg.Dialer)),
		cfg:              cfg,
		stopped:          make(chan interface{}),
		pingReset:        make(chan struct{}),
		outgoingPongs:    make(chan struct{}),
		outgoingPings:    make(chan struct{}),
		incomingMessages: make(chan *IncomingMessage, 100),
		decode:           decode,
		outgoingMessages: make(chan *outgoingMessage),
		notifyErrors:     make(chan error, 1),
	}, nil
}

// Start the client
// It establishes a connection to the server and starts the loops to handle incoming and
// outgoing messages
func (c *Client) Start(ctx context.Context) error {
	// connect to the server
	err := c.connect(ctx)
	if err != nil {
		return err
	}

	// Register control message handlers
	c.conn.SetPongHandler(c.handlePong)
	c.conn.SetPingHandler(c.handlePing)
	c.conn.SetCloseHandler(c.handleClose)

	// Start all the loops
	c.wg.Add(3)
	go func() {
		c.pingLoop()
		c.wg.Done()
	}()
	go func() {
		c.incomingMessageLoop()
		c.wg.Done()
	}()
	go func() {
		c.outgoingMessageLoop()
		c.wg.Done()
	}()

	return nil
}

// connect establishes a connection to the server
func (c *Client) connect(ctx context.Context) error {
	conn, _, err := c.dialer.DialContext(ctx, "", nil)
	c.conn = conn
	return err
}

// sendMessage schedules a message to be sent to the server
// sendMessage is non-blocking
// MUST not be called after calling close()
func (c *Client) SendMessage(ctx context.Context, msgType int, encode func(io.Writer) error) error {
	select {
	case <-c.stopped:
		return c.incomingErr
	default:
		msg := &outgoingMessage{
			ctx:     ctx,
			msgType: msgType,
			encode:  encode,
			err:     make(chan error),
		}
		defer close(msg.err)
		c.outgoingMessages <- msg
		err := <-msg.err
		return err
	}
}

// Messages returns a channel that receives incoming messages
func (c *Client) Messages() <-chan *IncomingMessage {
	return c.incomingMessages
}

// Errors notifies the users of any errors that occurs during the connection
// Users SHOULD listen to this channel to be notified of any errors that occurs indicating that the connection is no longer usable
// After receiving an error, the Client is no longer usable and the user MUST call Stop to ensure that the connection is properly closed
// and resources are released
func (c *Client) Errors() <-chan error {
	return c.notifyErrors
}

// IncomingMessage represents a message received from the server
type IncomingMessage struct {
	msgType int
	val     interface{}
	err     error
}

func (m *IncomingMessage) MsgType() int {
	return m.msgType
}

func (m *IncomingMessage) Value() interface{} {
	return m.val
}

func (m *IncomingMessage) Err() error {
	return m.err
}

// nextIncomingMessage wait for one message and puts it to the incoming channel
func (c *Client) incomingMessageLoop() {
	for {
		// Get the next reader from the connection
		msgType, r, err := c.conn.NextReader()
		if err != nil {
			c.incomingErr = err
			c.notifyErrors <- err
			if closeErr, ok := err.(*websocket.CloseError); ok {
				// We received a close message from the server
				c.serverStop(closeErr.Code)
			} else {
				// We received an error which is not a close message
				// We stop the connection with an abnormal closure
				c.stop(websocket.CloseAbnormalClosure)
			}

			return
		}

		// We secussfully read a message, so we reset the read deadline
		c.resetReadDeadline(c.cfg.ReadTimeout)

		if msgType != websocket.BinaryMessage && msgType != websocket.TextMessage {
			c.incomingErr = fmt.Errorf("unsupported message type %v", msgType)
			c.notifyErrors <- c.incomingErr
			c.stop(websocket.CloseUnsupportedData)
			return
		}

		// Decode the message
		val, err := c.decode(r)

		c.incomingMessages <- &IncomingMessage{msgType: msgType, val: val, err: err}
	}
}

type outgoingMessage struct {
	ctx     context.Context
	msgType int
	encode  func(io.Writer) error
	err     chan error
}

// outgoingMessageLoop sends messages to the server
// it is the only goroutine that writes to the connection
func (c *Client) outgoingMessageLoop() {
	for {
		select {
		case <-c.stopped:
			// Notify the server that we are closing down
			c.writeCloseMessage(websocket.CloseNormalClosure)

			// Drain the outgoing messages
			for msg := range c.outgoingMessages {
				msg.err <- c.incomingErr
			}
			return
		case <-c.outgoingPongs:
			c.writeControlMessage(websocket.PongMessage, nil)
		case <-c.outgoingPings:
			c.writeControlMessage(websocket.PingMessage, nil)
			c.resetReadDeadline(c.cfg.PongTimeout)
		case msg, ok := <-c.outgoingMessages:
			if ok {
				select {
				case <-msg.ctx.Done():
					// Context is done, so we don't attempt to write the message
					msg.err <- msg.ctx.Err()
				default:
					c.writeMessage(msg)
				}
			}
		}
	}
}

// writeMessage writes a message to the connection
func (c *Client) writeMessage(msg *outgoingMessage) {
	// Set the write deadline to the minimum of the context deadline and
	// the connection write deadline
	writeDeadline := time.Now().Add(c.cfg.WriteTimeout)
	deadline, ok := msg.ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.cfg.WriteTimeout)
	} else if deadline.After(writeDeadline) {
		deadline = writeDeadline
	}
	c.setWriteDeadline(deadline)

	// Get the next writer to the connection
	writer, err := c.conn.NextWriter(msg.msgType)
	if err != nil {
		msg.err <- err
		return
	}

	// Write the message to the connection
	msg.err <- msg.encode(writer)

	_ = writer.Close()
}

// handlePong is called when a pong message is received
func (c *Client) handlePong(_ string) error {
	// We secussfully read a message, so we reset the read deadline
	c.resetReadDeadline(c.cfg.ReadTimeout)

	c.pingReset <- struct{}{}
	return nil
}

// handlePing is called when a ping message is received
func (c *Client) handlePing(_ string) error {
	// We secussfully read a message, so we reset the read deadline
	c.resetReadDeadline(c.cfg.ReadTimeout)

	select {
	case c.outgoingPongs <- struct{}{}:
	default:
		// does not accumulate outgoing pongs
	}
	return nil
}

// handleClose is called when a close message is received
func (c *Client) handleClose(code int, _ string) error {
	// We secussfully read a message, so we reset the read deadline
	c.resetReadDeadline(c.cfg.ReadTimeout)

	c.serverStop(code)
	return nil
}

// pingLoop sends periodic ping frames when the connection is idle.
func (c *Client) pingLoop() {
	var pingTimer = time.NewTimer(c.cfg.PingInterval)
	defer pingTimer.Stop()

	for {
		select {
		case <-c.stopped:
			return
		case <-c.pingReset:
			if !pingTimer.Stop() {
				<-pingTimer.C
			}
			pingTimer.Reset(c.cfg.PingInterval)
		case <-pingTimer.C:
			c.outgoingPings <- struct{}{}
			pingTimer.Reset(c.cfg.PingInterval)
		}
	}
}

func (c *Client) writeCloseMessage(closeCode int) {
	fmt.Printf("writeCloseMessage: %v\n", closeCode)
	c.writeControlMessage(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, ""))
}

func (c *Client) writeControlMessage(msgType int, data []byte) {
	_ = c.conn.WriteControl(msgType, data, time.Now().Add(c.cfg.WriteControlTimeout))
}

func (c *Client) setWriteDeadline(deadline time.Time) {
	_ = c.conn.SetWriteDeadline(deadline)
}

func (c *Client) resetReadDeadline(timeout time.Duration) {
	if timeout > 0 {
		_ = c.conn.SetReadDeadline(time.Now().Add(timeout))
	}
}

// Stop stops the connection
//
// It attempts to gracefully stop the connection by sending a Websocket Close message to the server and waiting for
// the server to respond with a Close message in return and finally closes the underlying connection.
//
// If the context is canceled before, the underlying connection is abruptly closed and stop returns as soon as its
// internal goroutines have finished and the its state is cleaned up.
//
// Users SHOULD call Stop to ensure that the connection is properly closed and resources are released.
// Users MUST NOT call SendMessage after calling Stop.
func (c *Client) Stop(ctx context.Context) error {
	c.closeOnce.Do(func() {
		c.userStop()
		close(c.outgoingMessages)
		c.closeErr = c.waitLoops(ctx)
		c.clean()
	})
	return c.closeErr
}

// userStop is called when the user calls Stop
func (c *Client) userStop() {
	c.stop(websocket.CloseNormalClosure)
}

// serverStop is called when we receive an incoming close message or error from the server
func (c *Client) serverStop(closeMsgCode int) {
	c.stop(closeMsgCode)
}

// stop stops the connection
// closeMsgCode is the close message code to send to the server
func (c *Client) stop(closeMsgCode int) {
	c.stopOnce.Do(func() {
		c.closeMsgCode = closeMsgCode
		close(c.stopped)
	})
}

// waitLoops waits for the loops to finish
// if the context is canceled, it forces the connection to close
// if the connection is closed unexpectedly, it returns the error
func (c *Client) waitLoops(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Loops have properly finished
		c.conn.Close()
	case <-ctx.Done():
		// Context timed out, before loops properly finished, something might go wrong with conn closing
		// so we force the connection to close
		c.conn.Close()
		<-done
	}

	if websocket.IsUnexpectedCloseError(c.incomingErr, websocket.CloseNormalClosure) {
		return c.incomingErr
	}

	return nil
}

// clean cleans up the client
func (c *Client) clean() {
	close(c.notifyErrors)
	close(c.incomingMessages)
	close(c.pingReset)
	close(c.outgoingPongs)
	close(c.outgoingPings)
}
