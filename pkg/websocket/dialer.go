package websocket

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/gorilla/websocket"
	comnet "github.com/kkrt-labs/kakarot-controller/pkg/net"
)

// DialerConfig is a configuration for a websocket.Dialer.
type DialerConfig struct {
	Dialer *comnet.DialerConfig

	HandshakeTimeout                time.Duration // HandshakeTimeout specifies the duration for the handshake to complete.
	ReadBufferSize, WriteBufferSize int           // ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes.
	MessageSizeLimit                *int64        // nil default, 0 no limit
	EnableCompression               bool          // EnableCompression specifies if the client should attempt to negotiate per message compression (RFC 7692)

	ReadLimit int64       // ReadLimit specifies the maximum size in bytes for a message read from the peer. If a message exceeds the limit, the connection sends a close message to the peer and returns ErrReadLimit.
	Header    http.Header // Custom headers to be attached to the request opening the WebSocket connection
}

// SetDefault sets default values for DialerConfig
func (cfg *DialerConfig) SetDefault() *DialerConfig {
	if cfg.Dialer == nil {
		cfg.Dialer = new(comnet.DialerConfig)
	}
	cfg.Dialer.SetDefault()

	if cfg.ReadBufferSize == 0 {
		cfg.ReadBufferSize = 1024
	}

	if cfg.WriteBufferSize == 0 {
		cfg.WriteBufferSize = 1024
	}

	if cfg.HandshakeTimeout == 0 {
		cfg.HandshakeTimeout = 30 * time.Second
	}

	return cfg
}

// NewDialer creates a new websocket.Dialer with the given configuration.
func NewDialer(cfg *DialerConfig) (dialer Dialer) {
	dialer = &websocket.Dialer{
		NetDialContext:    comnet.NewDialer(cfg.Dialer).DialContext,
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
		HandshakeTimeout:  cfg.HandshakeTimeout,
		EnableCompression: cfg.EnableCompression,
		Proxy:             http.ProxyFromEnvironment,
	}

	dialer = WithError()(dialer)
	dialer = WithHeaders(cfg.Header)(dialer)
	dialer = WithReadLimit(cfg.ReadLimit)(dialer)

	return dialer
}

// Dialer is an interface for dialing a websocket connection.
type Dialer interface {
	DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)
}

// DialerDecorator is a function that decorates a Dialer.
type DialerDecorator func(Dialer) Dialer

// DialerFunc is a function that implements the Dialer interface.
type DialerFunc func(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)

// DialContext dials a websocket connection.
func (f DialerFunc) DialContext(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
	return f(ctx, urlStr, requestHeader)
}

// WithBaseURL is a DialerDecorator that sets the base URL for the websocket connection.
func WithBaseURL(baseURL *url.URL) DialerDecorator {
	return func(d Dialer) Dialer {
		return DialerFunc(func(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
			if urlStr == "" {
				u := *baseURL // copy the base URL
				if u.User != nil {
					b64auth := base64.StdEncoding.EncodeToString([]byte(u.User.String()))
					requestHeader.Add("authorization", "Basic "+b64auth)
					u.User = nil
				}

				urlStr = u.String()
			}

			return d.DialContext(ctx, urlStr, requestHeader)
		})
	}
}

// WithHeaders is a DialerDecorator that sets the headers for the websocket connection.
func WithHeaders(headers http.Header) DialerDecorator {
	return func(d Dialer) Dialer {
		return DialerFunc(func(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
			for key, values := range headers {
				requestHeader[key] = values
			}

			return d.DialContext(ctx, urlStr, requestHeader)
		})
	}
}

// WithAuth is a DialerDecorator that wraps a websocket dialing error in an expresive error
func WithError() DialerDecorator {
	return func(d Dialer) Dialer {
		return DialerFunc(func(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
			conn, resp, err := d.DialContext(ctx, urlStr, requestHeader)
			if err != nil {
				return conn, resp, autorest.NewErrorWithError(err, "jsonrpcws.Dialer", "DialContext", resp, "Dial")
			}
			return conn, resp, nil
		})
	}
}

// WithReadLimit is a DialerDecorator that sets the read limit for the websocket connection.
func WithReadLimit(limit int64) DialerDecorator {
	return func(d Dialer) Dialer {
		return DialerFunc(func(ctx context.Context, urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
			conn, resp, err := d.DialContext(ctx, urlStr, requestHeader)
			if err != nil {
				return conn, resp, err
			}

			conn.SetReadLimit(limit)

			return conn, resp, nil
		})
	}
}
