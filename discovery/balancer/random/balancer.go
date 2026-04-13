package random

import (
	"context"
	"math/rand/v2"
	"sync"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/balancer/simple"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

// Rand is the interface consumed by the random picker, allowing tests to
// inject a deterministic source.
type Rand interface {
	IntN(int) int
}

// lockedRand wraps a *rand.Rand with a mutex so it is safe for concurrent use.
type lockedRand struct {
	mu sync.Mutex
	r  *rand.Rand
}

func (lr *lockedRand) IntN(n int) int {
	lr.mu.Lock()
	v := lr.r.IntN(n)
	lr.mu.Unlock()

	return v
}

type picker[T peer.Peer] struct {
	rand Rand
}

func (p *picker[T]) Pick(_ context.Context, peers []T) (T, error) {
	return peers[p.rand.IntN(len(peers))], nil
}

func NewPolicy[T peer.Peer]() balancer.Policy[T] {
	return simple.NewPolicy(&picker[T]{
		rand: &lockedRand{r: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}, //nolint:gosec
	})
}

func NewBalancer[T peer.Peer](r resolver.Resolver[T]) balancer.Balancer[T] {
	return balancer.WrapPolicy(r, NewPolicy[T](), func(p T) (T, error) { return p, nil })
}
