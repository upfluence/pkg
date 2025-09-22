package random

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

type Rand interface {
	Intn(int) int
}

type Balancer[T peer.Peer] struct {
	*resolver.Puller[T]

	peers   []T
	peersMu *sync.RWMutex
	rand    Rand

	notifier chan interface{}
	closeFn  func()
}

func NewBalancer[T peer.Peer](r resolver.Resolver[T]) *Balancer[T] {
	var b = &Balancer[T]{
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
		peersMu:  &sync.RWMutex{},
		notifier: make(chan interface{}),
	}

	b.Puller, b.closeFn = resolver.NewPuller(r, b.updatePeers)

	return b
}

func (b *Balancer[T]) String() string {
	return fmt.Sprintf("loadbalancer/random [resolver: %v]", b.Puller)
}

func (b *Balancer[T]) updatePeers(u resolver.Update[T]) {
	b.peersMu.Lock()
	defer b.peersMu.Unlock()

	var newPeers = make(map[T]interface{})

	for _, p := range b.peers {
		var found bool

		for _, peer := range u.Deletions {
			if p.Addr() == peer.Addr() {
				found = true
			}
		}

		if !found {
			newPeers[p] = nil
		}
	}

	for _, p := range u.Additions {
		var found bool

		for _, peer := range b.peers {
			if p.Addr() == peer.Addr() {
				found = true
			}
		}

		if !found {
			newPeers[p] = nil
		}
	}

	var (
		i     = 0
		empty = len(b.peers) == 0
	)

	b.peers = make([]T, len(newPeers))

	for p, _ := range newPeers {
		b.peers[i] = p
		i++
	}

	if empty && (len(b.peers) > 0) {
		for {
			select {
			case <-b.notifier:
			default:
				return
			}
		}
	}
}

func (b *Balancer[T]) hasPeers() bool {
	b.peersMu.RLock()
	defer b.peersMu.RUnlock()

	return len(b.peers) > 0
}

func (b *Balancer[T]) Get(ctx context.Context, opts balancer.GetOptions) (T, func(error), error) {
	var zero T

	if !b.hasPeers() {
		if opts.NoWait {
			return zero, nil, balancer.ErrNoPeerAvailable
		}

		select {
		case b.notifier <- true:
		case <-ctx.Done():
			return zero, nil, ctx.Err()
		}
	}

	b.peersMu.RLock()
	defer b.peersMu.RUnlock()
	return b.peers[b.rand.Intn(len(b.peers))], func(error) {}, nil
}

func (b *Balancer[T]) Close() error {
	b.closeFn()
	return nil
}
