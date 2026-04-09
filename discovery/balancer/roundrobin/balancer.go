package roundrobin

import (
	"context"
	"sync/atomic"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/balancer/simple"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

type picker[T peer.Peer] struct {
	index atomic.Uint64
}

func (p *picker[T]) Pick(ctx context.Context, peers []T) (T, error) {
	idx := p.index.Add(1) - 1
	return peers[idx%uint64(len(peers))], nil
}

func NewPolicy[T peer.Peer]() balancer.Policy[T] {
	return simple.NewPolicy(&picker[T]{})
}

func NewBalancer[T peer.Peer](r resolver.Resolver[T]) balancer.Balancer[T] {
	return balancer.PolicyBalancerFunc(NewPolicy[T]())(r)
}
