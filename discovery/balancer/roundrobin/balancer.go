package roundrobin

import (
	"container/ring"
	"context"
	"sync"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

type Policy[T peer.Peer] struct {
	mu       sync.Mutex
	addrs    map[string]*ring.Ring
	ring     *ring.Ring
	notifier chan struct{}
}

func NewPolicy[T peer.Peer]() *Policy[T] {
	return &Policy[T]{
		addrs:    make(map[string]*ring.Ring),
		notifier: make(chan struct{}),
	}
}

func (p *Policy[T]) Update(u resolver.Update[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()

	wasEmpty := p.ring == nil

	for _, peer := range u.Additions {
		r := &ring.Ring{Value: peer}
		p.addrs[peer.Addr()] = r

		if p.ring == nil {
			p.ring = r
			continue
		}

		p.ring.Link(r)
	}

	for _, peer := range u.Deletions {
		addr := peer.Addr()
		r, ok := p.addrs[addr]

		if !ok {
			continue
		}

		delete(p.addrs, addr)

		// If this is the only element in the ring, set ring to nil
		if r.Len() == 1 {
			p.ring = nil
			continue
		}

		// Otherwise, unlink this element from the ring
		// Unlink returns the removed subring, so we keep prev as the new ring position
		prev := r.Prev()
		prev.Unlink(1) // Remove r from the ring
		p.ring = prev
	}

	isEmpty := p.ring == nil

	if wasEmpty && !isEmpty {
		close(p.notifier)
		p.notifier = make(chan struct{})
	} else if !wasEmpty && isEmpty {
		p.notifier = make(chan struct{})
	}
}

func (p *Policy[T]) Get(ctx context.Context, opts balancer.GetOptions) (T, func(error), error) {
	var zero T

	p.mu.Lock()
	r := p.ring
	notifier := p.notifier
	p.mu.Unlock()

	if r == nil {
		if opts.NoWait {
			return zero, nil, balancer.ErrNoPeerAvailable
		}

		select {
		case <-notifier:
		case <-ctx.Done():
			return zero, nil, ctx.Err()
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ring == nil {
		return zero, nil, balancer.ErrNoPeerAvailable
	}

	if v := p.ring.Value; v != nil {
		p.ring = p.ring.Next()
		return v.(T), func(error) {}, nil
	}

	return zero, nil, balancer.ErrNoPeerAvailable
}

func NewBalancer[T peer.Peer](r resolver.Resolver[T]) balancer.Balancer[T] {
	return balancer.PolicyBalancerFunc(NewPolicy[T]())(r)
}
