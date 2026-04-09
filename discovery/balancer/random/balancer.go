package random

import (
	"context"
	"math/rand"
	"time"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/balancer/simple"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

type Rand interface {
	Intn(int) int
}

type picker[T peer.Peer] struct {
	rand Rand
}

func (p *picker[T]) Pick(ctx context.Context, peers []T) (T, error) {
	return peers[p.rand.Intn(len(peers))], nil
}

func NewPolicy[T peer.Peer]() balancer.Policy[T] {
	return simple.NewPolicy(&picker[T]{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	})
}

func NewBalancer[T peer.Peer](r resolver.Resolver[T]) balancer.Balancer[T] {
	return balancer.WrapPolicy(r, NewPolicy[T](), func(p T) (T, error) { return p, nil })
}
