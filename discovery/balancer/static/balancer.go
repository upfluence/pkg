package static

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/discovery/resolver"
)

var errNoPeer = errors.New("balancer/static: No Peer available")

type Balancer struct {
	resolver resolver.Resolver

	peer *peer.Peer
	mu   *sync.RWMutex
}

func NewBalancer(r resolver.Resolver) *Balancer {
	return &Balancer{resolver: r, mu: &sync.RWMutex{}}
}

func (b *Balancer) String() string {
	return fmt.Sprintf("loadbalancer/random [resolver: %v]", b.resolver)
}

func (*Balancer) Open(context.Context) error { return nil }
func (*Balancer) IsOpen() bool               { return true }
func (*Balancer) Close() error               { return nil }

func (b *Balancer) Get(ctx context.Context, _ balancer.BalancerGetOptions) (*peer.Peer, error) {
	if p := b.getPeer(); p != nil {
		return p, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.resolver.Open(ctx); err != nil {
		return nil, err
	}

	defer b.resolver.Close()

	updates, err := b.resolver.Resolve(ctx)

	if err != nil {
		return nil, err
	}

	select {
	case u := <-updates:
		if len(u.Additions) == 0 {
			return nil, errNoPeer
		}

		return u.Additions[0], nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (b *Balancer) getPeer() *peer.Peer {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.peer
}
