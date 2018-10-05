package runtimevar

import (
	"context"

	"github.com/google/go-cloud/runtimevar/driver"
	"github.com/micro/go-config/source"
)

type driverWatcherKey struct{}

// WithWatcher sets the runtimevar driver.Watcher
func WithWatcher(dv driver.Watcher) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, driverWatcherKey{}, dv)
	}
}
