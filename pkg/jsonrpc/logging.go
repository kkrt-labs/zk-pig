package jsonrpc

import (
	"context"

	"github.com/kkrt-labs/kakarot-controller/pkg/log"
	"github.com/kkrt-labs/kakarot-controller/pkg/tag"
)

// WithTags is a decorator that attaches JSON-RPC specific tags to the provided context namespaces.
// If no namespaces are provided, the tags are attached to the default namespace.

// It attaches the following tags:
// - req.method: JSON-RPC method
// - req.version: JSON-RPC version
// - req.params: JSON-RPC params
// - req.id: JSON-RPC id

// It also attaches the provided component name to the default component tag.
func WithTags(component string, namespaces ...string) ClientDecorator {
	return func(c Client) Client {
		return ClientFunc(func(ctx context.Context, req *Request, res interface{}) error {
			if component != "" {
				ctx = tag.WithComponent(ctx, component)
			}

			tags := []*tag.Tag{
				tag.Key("req.method").String(req.Method),
				tag.Key("req.version").String(req.Version),
				tag.Key("req.params").Object(req.Params),
				tag.Key("req.id").Object(req.ID),
			}

			if len(namespaces) == 0 {
				namespaces = []string{tag.DefaultNamespace}
			}

			for _, ns := range namespaces {
				ctx = tag.WithNamespaceTags(ctx, ns, tags...)
			}

			return c.Call(ctx, req, res)
		})
	}
}

func WithLog(namespaces ...string) ClientDecorator {
	return func(c Client) Client {
		return ClientFunc(func(ctx context.Context, req *Request, res interface{}) error {
			logger := log.SugaredLoggerWithFieldsFromNamespaceContext(ctx, namespaces...)

			logger.Debug("Call JSON-RPC")
			err := c.Call(ctx, req, res)
			if err != nil {
				logger.Errorf("JSON-RPC call failed with error: %v", err)
			}

			return err
		})
	}
}
