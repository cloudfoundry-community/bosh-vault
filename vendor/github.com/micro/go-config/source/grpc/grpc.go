package grpc

import (
	"context"

	"github.com/micro/go-config/source"
	proto "github.com/micro/go-config/source/grpc/proto"
	"google.golang.org/grpc"
)

type grpcSource struct {
	addr string
	path string
	opts source.Options
}

var (
	DefaultPath    = "/micro/config"
	DefaultAddress = "localhost:8080"
)

func (g *grpcSource) Read() (*source.ChangeSet, error) {
	c, err := grpc.Dial(g.addr)
	if err != nil {
		return nil, err
	}
	cl := proto.NewSourceClient(c)
	rsp, err := cl.Read(context.Background(), &proto.ReadRequest{
		Path: g.path,
	})
	if err != nil {
		return nil, err
	}
	return toChangeSet(rsp.ChangeSet), nil
}

func (g *grpcSource) Watch() (source.Watcher, error) {
	c, err := grpc.Dial(g.addr)
	if err != nil {
		return nil, err
	}
	cl := proto.NewSourceClient(c)
	rsp, err := cl.Watch(context.Background(), &proto.WatchRequest{
		Path: g.path,
	})
	if err != nil {
		return nil, err
	}
	return newWatcher(rsp)
}

func (g *grpcSource) String() string {
	return "grpc"
}

func NewSource(opts ...source.Option) source.Source {
	var options source.Options
	for _, o := range opts {
		o(&options)
	}

	addr := DefaultAddress
	path := DefaultPath

	if options.Context != nil {
		a, ok := options.Context.Value(addressKey{}).(string)
		if ok {
			addr = a
		}
		p, ok := options.Context.Value(pathKey{}).(string)
		if ok {
			path = p
		}
	}

	return &grpcSource{
		addr: addr,
		path: path,
		opts: options,
	}
}
