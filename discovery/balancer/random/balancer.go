package random

import (
	"context"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

type Rand interface {
	Intn(int) int
}

type Policy[T peer.Peer] struct {
	mu       sync.RWMutex
	peers    []T
	rand     Rand
	notifier chan struct{}
}

func NewPolicy[T peer.Peer]() *Policy[T] {
	return &Policy[T]{
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
		notifier: make(chan struct{}),
	}
}

func (p *Policy[T]) Update(u resolver.Update[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()

	wasEmpty := len(p.peers) == 0

	peerMap := make(map[string]T)
	for _, peer := range p.peers {
		peerMap[peer.Addr()] = peer
	}

	for _, peer := range u.Deletions {
		delete(peerMap, peer.Addr())
	}

	for _, peer := range u.Additions {
		peerMap[peer.Addr()] = peer
	}

	p.peers = make([]T, 0, len(peerMap))
	for _, peer := range peerMap {
		p.peers = append(p.peers, peer)
	}

	if wasEmpty && len(p.peers) > 0 {
		close(p.notifier)
		p.notifier = make(chan struct{})
	}
}

func (p *Policy[T]) Get(ctx context.Context, opts balancer.GetOptions) (T, func(error), error) {
	var zero T

	p.mu.RLock()
	hasPeers := len(p.peers) > 0
	notifier := p.notifier
	peers := slices.Clone(p.peers)
	p.mu.RUnlock()

	if !hasPeers {
		if opts.NoWait {
			return zero, nil, balancer.ErrNoPeerAvailable
		}

		select {
		case <-notifier:
		case <-ctx.Done():
			return zero, nil, ctx.Err()
		}

		p.mu.RLock()
		peers = slices.Clone(p.peers)
		p.mu.RUnlock()
	}

	if len(peers) == 0 {
		return zero, nil, balancer.ErrNoPeerAvailable
	}

	return peers[p.rand.Intn(len(peers))], func(error) {}, nil
}

func NewBalancer[T peer.Peer](r resolver.Resolver[T]) balancer.Balancer[T] {
	return balancer.WrapPolicy(r, NewPolicy[T](), func(p T) (T, error) { return p, nil })
}
