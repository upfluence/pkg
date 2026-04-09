package simple

import (
	"context"
	"slices"
	"sync"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

// Picker selects a peer from a list of available peers.
type Picker[T peer.Peer] interface {
	// Pick selects a peer from the provided list. The implementation should
	// return an error if no suitable peer can be selected.
	Pick(context.Context, []T) (T, error)
}

type policy[T peer.Peer] struct {
	picker Picker[T]

	mu       sync.RWMutex
	peers    []T
	notifier chan struct{}
}

// NewPolicy creates a new Policy that delegates peer selection to the provided Picker.
func NewPolicy[T peer.Peer](picker Picker[T]) balancer.Policy[T] {
	return &policy[T]{
		picker:   picker,
		notifier: make(chan struct{}),
	}
}

func (p *policy[T]) Update(u resolver.Update[T]) {
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

func (p *policy[T]) Get(ctx context.Context, opts balancer.GetOptions) (T, func(error), error) {
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

	peer, err := p.picker.Pick(ctx, peers)
	return peer, func(error) {}, err
}
