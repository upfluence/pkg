package balancer

import (
	"context"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

var ErrNoPeerAvailable = errors.New("balancer: No peer available")

type GetOptions struct {
	NoWait bool
}

type Builder[T peer.Peer] interface {
	Build(string) Balancer[T]
}

type ResolverBuilder[T peer.Peer] struct {
	Builder      resolver.Builder[T]
	BalancerFunc func(resolver.Resolver[T]) Balancer[T]
}

func (rb ResolverBuilder[T]) Build(k string) Balancer[T] {
	return rb.BalancerFunc(rb.Builder.Build(k))
}

type Balancer[T peer.Peer] interface {
	Open(context.Context) error
	IsOpen() bool
	Close() error

	Get(context.Context, GetOptions) (T, func(error), error)
}
