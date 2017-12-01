package roundrobin

import (
	"container/ring"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/upfluence/pkg/backoff/static"
	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/discovery/resolver"
	"github.com/upfluence/pkg/log"
)

var getBackoffStrategy = static.NewInfiniteBackoff(100 * time.Millisecond)

type Balancer struct {
	resolver resolver.Resolver

	ring   *ring.Ring
	ringMu *sync.Mutex

	closeChan chan bool
}

func NewBalancer(r resolver.Resolver) *Balancer {
	return &Balancer{
		resolver:  r,
		ring:      &ring.Ring{},
		ringMu:    &sync.Mutex{},
		closeChan: make(chan bool),
	}
}

func (b *Balancer) String() string {
	return fmt.Sprintf("loadbalancer/roundrobin [resolver: %v]", b.resolver)
}

func (b *Balancer) Open(ctx context.Context) error {
	if err := b.resolver.Open(ctx); err != nil {
		return err
	}

	go b.subscribe()

	return nil
}

func (b *Balancer) subscribe() {
	for {
		var (
			channelOpen = true
			ch, err     = b.resolver.Resolve(context.Background())
		)

		if err != nil {
			log.Errorf("resolver: %+v", err)
		}

		for channelOpen {
			select {
			case <-b.closeChan:
				return
			case update, ok := <-ch:
				if !ok {
					channelOpen = false
				} else {
					b.updateRing(update)
				}
			}
		}
	}
}

func (b *Balancer) updateRing(update resolver.Update) {
	if len(update.Additions)+len(update.Deletions) == 0 {
		return
	}

	b.ringMu.Lock()
	defer b.ringMu.Unlock()

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
}

func (b *Balancer) Close() error {
	b.closeChan <- true

	return b.resolver.Close()
}

func (b *Balancer) getNoWait() *peer.Peer {
	b.ringMu.Lock()
	defer b.ringMu.Unlock()

	if v := b.ring.Value; v != nil {
		b.ring = b.ring.Next()

		p := v.(peer.Peer)
		return &p
	}

	return nil
}

func (b *Balancer) Get(ctx context.Context, opts balancer.BalancerGetOptions) (*peer.Peer, error) {
	if peer := b.getNoWait(); peer != nil {
		return peer, nil
	}

	if opts.NoWait {
		return nil, balancer.ErrNoPeerAvailable
	}

	nextPeer := make(chan *peer.Peer)

	go func() {
		var (
			i = 0
			p *peer.Peer
		)

		for p == nil {
			d, _ := getBackoffStrategy.Backoff(i)
			log.Infof(
				"[%v] No peer found, backing off for: %v",
				b,
				d,
			)

			time.Sleep(d)

			p = b.getNoWait()
			i++
		}

		nextPeer <- p
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-nextPeer:
		return p, nil
	}
}
