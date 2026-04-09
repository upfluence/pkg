package balancer

import (
	"context"
	"sync"

	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

type Policy[T peer.Peer] interface {
	Get(context.Context, GetOptions) (T, func(error), error)
	Update(resolver.Update[T])
}

type policyBalancer[S, T peer.Peer] struct {
	resolver.Puller[S]
	Policy[T]

	mu      sync.Mutex
	peers   map[string]T
	builder func(S) (T, error)
}

func WrapPolicy[S, T peer.Peer](r resolver.Resolver[S], p Policy[T], build func(S) (T, error)) Balancer[T] {
	b := &policyBalancer[S, T]{
		Policy:  p,
		peers:   make(map[string]T),
		builder: build,
	}

	b.Puller = resolver.Puller[S]{
		Resolver:   r,
		UpdateFunc: b.handleUpdate,
	}

	return b
}

func PolicyBalancerFunc[T peer.Peer](p Policy[T]) func(resolver.Resolver[T]) Balancer[T] {
	return func(r resolver.Resolver[T]) Balancer[T] {
		return WrapPolicy(r, p, func(t T) (T, error) { return t, nil })
	}
}

func (b *policyBalancer[S, T]) handleUpdate(u resolver.Update[S]) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var mapped resolver.Update[T]

	for _, sp := range u.Additions {
		tp, err := b.builder(sp)
		if err != nil {
			continue
		}

		b.peers[sp.Addr()] = tp
		mapped.Additions = append(mapped.Additions, tp)
	}

	for _, sp := range u.Deletions {
		if tp, ok := b.peers[sp.Addr()]; ok {
			delete(b.peers, sp.Addr())
			mapped.Deletions = append(mapped.Deletions, tp)
		}
	}

	b.Policy.Update(mapped)
}
