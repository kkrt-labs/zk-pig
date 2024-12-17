package jsonrpc

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/kkrt-labs/kakarot-controller/pkg/log"
	"go.uber.org/zap"
)

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

			return backoff.RetryNotify(
				func() error { return c.Call(ctx, req, res) },
				backoff.WithContext(bckff, ctx),
				func(err error, d time.Duration) {
					log.LoggerFromContext(ctx).Warn("JSON-RPC call failed retrying...", 
						zap.Error(err),
						zap.Duration("duration", d),
					)
				},
			)
		})
	}
}
