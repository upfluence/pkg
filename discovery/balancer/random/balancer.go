package random

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/discovery/resolver"
)

type Rand interface {
	Intn(int) int
}

type Balancer struct {
	*resolver.Puller

	peers   []*peer.Peer
	peersMu *sync.RWMutex
	rand    Rand

	notifier  chan interface{}
	closeChan chan<- interface{}
}

func NewBalancer(r resolver.Resolver) *Balancer {
	var b = &Balancer{
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
		peersMu:  &sync.RWMutex{},
		notifier: make(chan interface{}),
	}

	b.Puller, b.closeChan = resolver.NewPuller(r, b.updatePeers)

	return b
}

func (b *Balancer) String() string {
	return fmt.Sprintf("loadbalancer/random [resolver: %v]", b.Puller)
}

func (b *Balancer) updatePeers(u resolver.Update) {
	b.peersMu.Lock()
	defer b.peersMu.Unlock()

	var newPeers = make(map[*peer.Peer]interface{})

	for _, p := range b.peers {
		var found bool

		for _, peer := range u.Deletions {
			if p.Addr == peer.Addr {
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
			if p.Addr == peer.Addr {
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

	b.peers = make([]*peer.Peer, len(newPeers))

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

func (b *Balancer) hasPeers() bool {
	b.peersMu.RLock()
	defer b.peersMu.RUnlock()

	return len(b.peers) > 0
}

func (b *Balancer) Get(ctx context.Context, opts balancer.BalancerGetOptions) (*peer.Peer, error) {
	if !b.hasPeers() {
		if opts.NoWait {
			return nil, balancer.ErrNoPeerAvailable
		}

		select {
		case b.notifier <- true:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	b.peersMu.RLock()
	defer b.peersMu.RUnlock()
	return b.peers[b.rand.Intn(len(b.peers))], nil
}

func (b *Balancer) Close() error {
	close(b.closeChan)
	return nil
}
