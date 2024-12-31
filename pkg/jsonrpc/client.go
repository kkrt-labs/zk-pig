package jsonrpc

import "context"

//go:generate mockgen -source client.go -destination mock/client.go -package jsonrpcmock Client

type Client interface {
	// Call performs a JSON-RPC call with the given request and store result in res

	// res MUST be a pointer so JSON-RPC result can be unmarshalled into res. You
	// can also pass nil, in which case the result is ignored.
	Call(ctx context.Context, req *Request, res interface{}) error
}

type ClientFunc func(ctx context.Context, req *Request, res interface{}) error

func (f ClientFunc) Call(ctx context.Context, req *Request, res interface{}) error {
	return f(ctx, req, res)
}
