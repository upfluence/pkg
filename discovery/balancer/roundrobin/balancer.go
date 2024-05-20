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

type Balancer[T peer.Peer] struct {
	resolver.Puller[T]

	addrs  map[string]*ring.Ring
	ring   *ring.Ring
	ringMu sync.RWMutex

	notifier chan struct{}
}

func BalancerFunc[T peer.Peer](r resolver.Resolver[T]) balancer.Balancer[T] {
	return NewBalancer[T](r)
}

func NewBalancer[T peer.Peer](r resolver.Resolver[T]) *Balancer[T] {
	var b = Balancer[T]{
		addrs:    make(map[string]*ring.Ring),
		notifier: make(chan struct{}),
	}

	b.Puller = resolver.Puller[T]{Resolver: r, UpdateFunc: b.updateRing}

	return &b
}

func (b *Balancer[T]) String() string {
	return fmt.Sprintf("loadbalancer/roundrobin [resolver: %v]", &b.Puller)
}

func (b *Balancer[T]) updateRing(update resolver.Update[T]) {
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

func (b *Balancer[T]) Get(ctx context.Context, opts balancer.GetOptions) (T, func(error), error) {
	var zero T

	b.ringMu.RLock()
	r := b.ring
	n := b.notifier
	b.ringMu.RUnlock()

	if r == nil {
		if opts.NoWait {
			return zero, nil, balancer.ErrNoPeerAvailable
		}

		pctx := b.Puller.Monitor.Context()

		select {
		case <-n:
		case <-ctx.Done():
			return zero, nil, ctx.Err()
		case <-pctx.Done():
			return zero, nil, pctx.Err()
		}
	}

	b.ringMu.Lock()
	defer b.ringMu.Unlock()

	if v := b.ring.Value; v != nil {
		b.ring = b.ring.Next()

		return v.(T), func(error) {}, nil
	}

	return zero, nil, balancer.ErrNoPeerAvailable
}
