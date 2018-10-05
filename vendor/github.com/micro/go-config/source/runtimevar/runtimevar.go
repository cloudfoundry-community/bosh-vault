// package runtimevar is the source for github.com/google/go-cloud/runtimevar
package runtimevar

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/go-cloud/runtimevar"
	"github.com/google/go-cloud/runtimevar/driver"
	"github.com/micro/go-config/source"
)

type rvSource struct {
	opts source.Options

	sync.Mutex
	v  *runtimevar.Variable
	dv driver.Watcher
}

func (rv *rvSource) Read() (*source.ChangeSet, error) {
	s, err := rv.v.Watch(context.Background())
	if err != nil {
		return nil, err
	}

	// assuming value is bytes
	b, err := rv.opts.Encoder.Encode(s.Value.([]byte))
	if err != nil {
		return nil, fmt.Errorf("error reading source: %v", err)
	}

	cs := &source.ChangeSet{
		Timestamp: s.UpdateTime,
		Format:    rv.opts.Encoder.String(),
		Source:    rv.String(),
		Data:      b,
	}
	cs.Checksum = cs.Sum()

	return cs, nil
}

func (rv *rvSource) Watch() (source.Watcher, error) {
	return newWatcher(rv.String(), rv.dv, rv.opts)
}

func (rv *rvSource) String() string {
	return "runtimevar"
}

func NewSource(opts ...source.Option) source.Source {
	options := source.NewOptions(opts...)

	dv, ok := options.Context.Value(driverWatcherKey{}).(driver.Watcher)
	if !ok {
		// nooooooo
		panic("driver watcher required")
	}

	return &rvSource{
		opts: options,
		v:    runtimevar.New(dv),
	}
}
