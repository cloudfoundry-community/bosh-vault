package grpc

import (
	"context"

	"github.com/micro/go-config/source"
)

type addressKey struct{}
type pathKey struct{}

// WithAddress sets the consul address
func WithAddress(a string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, addressKey{}, a)
	}
}

// WithPath sets the key prefix to use
func WithPath(p string) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, pathKey{}, p)
	}
}
