package stapler

import (
	"fmt"
	"os"

	"github.com/upfluence/pkg/peer"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/uthrift-go/uthrift/context"
)

type UUIDGenerator interface {
	Generate() string
}

type randomGenerator struct {
	*os.File
}

func NewRandomGenerator() UUIDGenerator {
	f, _ := os.Open("/dev/urandom")

	return &randomGenerator{File: f}
}

func (g *randomGenerator) Generate() string {
	var b = make([]byte, 16)
	g.Read(b)

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type Factory struct {
	peer      *peer.Peer
	generator UUIDGenerator
}

func NewFactory(p *peer.Peer, g UUIDGenerator) *Factory {
	return &Factory{peer: p, generator: g}
}

func NewDefaultFactory() *Factory {
	return &Factory{peer: peer.FromEnv(), generator: NewRandomGenerator()}
}

func (f *Factory) GetMiddleware(string, string) thrift.TMiddleware {
	return &Middleware{peer: f.peer, generator: f.generator}
}

type Middleware struct {
	peer      *peer.Peer
	generator UUIDGenerator
}

func (m *Middleware) staple(ctx thrift.Context) thrift.Context {
	return context.WithSpanID(
		context.WithPeer(ctx, m.peer),
		m.generator.Generate(),
	)
}

func (m *Middleware) HandleBinaryRequest(ctx thrift.Context, _ string, _ int32, req thrift.TRequest, next func(thrift.Context, thrift.TRequest) (thrift.TResponse, error)) (thrift.TResponse, error) {
	return next(m.staple(ctx), req)
}

func (m *Middleware) HandleUnaryRequest(ctx thrift.Context, _ string, _ int32, req thrift.TRequest, next func(thrift.Context, thrift.TRequest) error) error {
	return next(m.staple(ctx), req)
}
