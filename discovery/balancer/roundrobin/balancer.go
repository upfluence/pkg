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
	puller *resolver.Puller

	ring   *ring.Ring
	ringMu *sync.Mutex

	closeFn  func()
	notifier chan interface{}
}

func NewBalancer(r resolver.Resolver) *Balancer {
	var b = &Balancer{
		ring:     &ring.Ring{},
		ringMu:   &sync.Mutex{},
		notifier: make(chan interface{}),
	}

	b.puller, b.closeFn = resolver.NewPuller(r, b.updateRing)

	return b
}

func (b *Balancer) String() string {
	return fmt.Sprintf("loadbalancer/roundrobin [resolver: %v]", b.puller)
}

func (b *Balancer) Open(ctx context.Context) error {
	return b.puller.Open(ctx)
}

func (b *Balancer) updateRing(update resolver.Update) {
	b.ringMu.Lock()
	defer b.ringMu.Unlock()

	var emptyRing = b.ring.Value == nil

	for _, p := range update.Additions {
		if b.ring.Value == nil {
			b.ring.Value = *p
		} else {
			b.ring.Link(&ring.Ring{Value: *p})
		}
	}

	for _, p := range update.Deletions {
		var (
			r     = b.ring
			found = false
		)

		b.ring = b.ring.Next()

		for !found || r != b.ring {
			if r.Value.(peer.Peer).Addr == p.Addr {
				found = true
			} else {
				b.ring = b.ring.Next()
			}
		}

		if found {
			b.ring = b.ring.Prev().Unlink(1)
		}
	}

	if emptyRing && b.ring.Value != nil {
		for {
			select {
			case <-b.notifier:
			default:
				return
			}
		}
	}
}

func (b *Balancer) Close() error {
	b.closeFn()
	return nil
}

func (b *Balancer) Get(ctx context.Context, opts balancer.BalancerGetOptions) (*peer.Peer, error) {
	if v := b.ring.Value; v == nil {
		if opts.NoWait {
			return nil, balancer.ErrNoPeerAvailable
		}

		select {
		case b.notifier <- true:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	b.ringMu.Lock()
	defer b.ringMu.Unlock()

	if v := b.ring.Value; v != nil {
		b.ring = b.ring.Next()

		p := v.(peer.Peer)
		return &p, nil
	}

	return nil, balancer.ErrNoPeerAvailable
}
