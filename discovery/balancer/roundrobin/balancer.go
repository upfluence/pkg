package roundrobin

import (
	"container/ring"
	"context"
	"fmt"
	"sync"

	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/discovery/resolver"
)

type Balancer struct {
	resolver.Puller

	addrs  map[string]*ring.Ring
	ring   *ring.Ring
	ringMu sync.RWMutex

	notifier chan struct{}
}

func BalancerFunc(r resolver.Resolver) balancer.Balancer {
	return NewBalancer(r)
}

func NewBalancer(r resolver.Resolver) *Balancer {
	var b = Balancer{
		addrs:    make(map[string]*ring.Ring),
		notifier: make(chan struct{}),
	}

	b.Puller = resolver.Puller{Resolver: r, UpdateFunc: b.updateRing}

	return &b
}

func (b *Balancer) String() string {
	return fmt.Sprintf("loadbalancer/roundrobin [resolver: %v]", &b.Puller)
}

func (b *Balancer) updateRing(update resolver.Update) {
	b.ringMu.Lock()
	defer b.ringMu.Unlock()

	wasEmpty := b.ring == nil

	for _, p := range update.Additions {
		r := &ring.Ring{Value: p}
		b.addrs[p.Addr()] = r

		if b.ring == nil {
			b.ring = r
			continue
		}

		b.ring.Link(r)
	}

	for _, p := range update.Deletions {
		addr := p.Addr()
		r, ok := b.addrs[addr]

		if !ok {
			continue
		}

		delete(b.addrs, addr)

		if p := r.Prev(); p != nil {
			b.ring = p.Unlink(1)
			continue
		}

		b.ring = nil
	}

	isEmpty := b.ring == nil

	if wasEmpty && !isEmpty {
		close(b.notifier)
	} else if !wasEmpty && isEmpty {
		b.notifier = make(chan struct{})
	}
}

func (b *Balancer) Get(ctx context.Context, opts balancer.GetOptions) (peer.Peer, error) {
	b.ringMu.RLock()
	r := b.ring
	n := b.notifier
	b.ringMu.RUnlock()

	if r == nil {
		if opts.NoWait {
			return nil, balancer.ErrNoPeerAvailable
		}

		pctx := b.Puller.Monitor.Context()

		select {
		case <-n:
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-pctx.Done():
			return nil, pctx.Err()
		}
	}

	b.ringMu.Lock()
	defer b.ringMu.Unlock()

	if v := b.ring.Value; v != nil {
		b.ring = b.ring.Next()

		p := v.(peer.Peer)
		return p, nil
	}

	return nil, balancer.ErrNoPeerAvailable
}
