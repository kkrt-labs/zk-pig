package jsonrpc

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/kkrt-labs/kakarot-controller/pkg/log"
	"go.uber.org/zap"
)

// ClientDecorator is a function that enable to decorate a JSON-RPC client with additional functionality
type ClientDecorator func(Client) Client

// WithVersion automatically set JSON-RPC request version
func WithVersion(v string) ClientDecorator {
	return func(c Client) Client {
		return ClientFunc(func(ctx context.Context, req *Request, res interface{}) error {
			req.Version = v
			return c.Call(ctx, req, res)
		})
	}
}

// WithIncrementalID automatically increments JSON-RPC request ID
func WithIncrementalID() ClientDecorator {
	var idCounter uint32
	return func(c Client) Client {
		return ClientFunc(func(ctx context.Context, req *Request, res interface{}) error {
			req.ID = atomic.AddUint32(&idCounter, 1) - 1
			return c.Call(ctx, req, res)
		})
	}
}

// WithRetry automatically retries JSON-RPC calls
func WithRetry() ClientDecorator {
	pool := &sync.Pool{
		New: func() interface{} {
			return backoff.NewExponentialBackOff(
				backoff.WithInitialInterval(50*time.Millisecond),
				backoff.WithMaxElapsedTime(2*time.Second),
			)
		},
	}
	return func(c Client) Client {
		return ClientFunc(func(ctx context.Context, req *Request, res interface{}) error {
			bckff := pool.Get().(*backoff.ExponentialBackOff)
			defer func() {
				bckff.Reset()
				pool.Put(bckff)
			}()

			attempt := 0
			attemptReq := req
			return backoff.RetryNotify(
				func() error {
					return c.Call(ctx, attemptReq, res)
				},
				backoff.WithContext(bckff, ctx),
				func(err error, d time.Duration) {
					attempt++
					// We need to increment the ID for each retry attempt
					// so that we don't possibly overwrite the response of the previous attempt
					attemptReq = &Request{
						Method:  req.Method,
						Version: req.Version,
						Params:  req.Params,
						ID:      fmt.Sprintf("%s#%d", req.ID, attempt),
					}
					log.LoggerFromContext(ctx).Warn("Retrying in...",
						zap.Error(err),
						zap.Duration("duration", d),
					)
				},
			)
		})
	}
}

// WithTimeout automatically sets a timeout for JSON-RPC calls
func WithTimeout(d time.Duration) ClientDecorator {
	return func(c Client) Client {
		return ClientFunc(func(ctx context.Context, req *Request, res interface{}) error {
			ctx, cancel := context.WithTimeout(ctx, d)
			defer cancel()

			return c.Call(ctx, req, res)
		})
	}
}
