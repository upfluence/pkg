package balancer

import (
	"context"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/discovery/resolver"
)

var ErrNoPeerAvailable = errors.New("balancer: No peer available")

type GetOptions struct {
	NoWait bool
}

type Builder interface {
	Build(string) Balancer
}

type ResolverBuilder struct {
	Builder      resolver.Builder
	BalancerFunc func(resolver.Resolver) Balancer
}

func (rb ResolverBuilder) Build(k string) Balancer {
	return rb.BalancerFunc(rb.Builder.Build(k))
}

type Balancer interface {
	Open(context.Context) error
	IsOpen() bool
	Close() error

	Get(context.Context, GetOptions) (peer.Peer, error)
}
