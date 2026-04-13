package static

import (
	"context"
	"fmt"
	"sync"

	"github.com/upfluence/pkg/v2/closer"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/metadata"
)

type Builder[T peer.Peer] map[string][]T

func (b Builder[T]) Build(n string) resolver.Resolver[T] {
	return NewResolver(b[n])
}

func PeersFromStrings(addrs ...string) []Peer {
	var peers = make([]Peer, len(addrs))

	for i, addr := range addrs {
		peers[i] = Peer(addr)
	}

	return peers
}

type Resolver[T peer.Peer] struct {
	closer.Monitor

	mu  sync.Mutex
	chs []chan resolver.Update[T]

	peers []T
}

type Peer string

func (sp Peer) Addr() string                { return string(sp) }
func (sp Peer) Metadata() metadata.Metadata { return nil }

func NewResolverFromStrings(addrs []string) *Resolver[Peer] {
	return NewResolver(PeersFromStrings(addrs...))
}

func NewResolver[T peer.Peer](peers []T) *Resolver[T] {
	return &Resolver[T]{peers: peers}
}

func (r *Resolver[T]) Peers() []T {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]T, len(r.peers))
	copy(out, r.peers)

	return out
}

func (r *Resolver[T]) UpdatePeers(peers []T) {
	r.mu.Lock()

	old := make(map[string]T, len(r.peers))
	for _, p := range r.peers {
		old[p.Addr()] = p
	}

	cur := make(map[string]T, len(peers))
	for _, p := range peers {
		cur[p.Addr()] = p
	}

	var u resolver.Update[T]

	for addr, p := range cur {
		if _, ok := old[addr]; !ok {
			u.Additions = append(u.Additions, p)
		}
	}

	for addr, p := range old {
		if _, ok := cur[addr]; !ok {
			u.Deletions = append(u.Deletions, p)
		}
	}

	r.peers = peers

	if len(u.Additions) == 0 && len(u.Deletions) == 0 {
		r.mu.Unlock()

		return
	}

	chs := make([]chan resolver.Update[T], len(r.chs))
	copy(chs, r.chs)
	r.mu.Unlock()

	for _, ch := range chs {
		// Non-blocking send: if the watcher's buffer is already full (slow
		// consumer), merge the pending update with the new one so no update
		// is silently dropped and the caller is never blocked.
		select {
		case ch <- u:
		default:
			select {
			case pending := <-ch:
				pending.Additions = append(pending.Additions, u.Additions...)

				pending.Deletions = append(pending.Deletions, u.Deletions...)
				ch <- pending
			default:
				// Channel was drained concurrently; just send the new update.
				ch <- u
			}
		}
	}
}

func (r *Resolver[T]) subscribe() (chan resolver.Update[T], []T) {
	ch := make(chan resolver.Update[T], 1)

	r.mu.Lock()
	r.chs = append(r.chs, ch)

	peers := make([]T, len(r.peers))
	copy(peers, r.peers)
	r.mu.Unlock()

	return ch, peers
}

func (r *Resolver[T]) unsubscribe(ch chan resolver.Update[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, c := range r.chs {
		if c == ch {
			r.chs = append(r.chs[:i], r.chs[i+1:]...)

			return
		}
	}
}

func (r *Resolver[T]) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var addrs = make([]string, len(r.peers))

	for i, peer := range r.peers {
		addrs[i] = peer.Addr()
	}

	return fmt.Sprintf("resolver/static: [seeds: %v]", addrs)
}

func (r *Resolver[T]) Build(string) resolver.Resolver[T] { return r }

func (r *Resolver[T]) Open(_ context.Context) error { return nil }

func (r *Resolver[T]) Resolve() resolver.Watcher[T] {
	return &watcher[T]{r: r}
}

type watcher[T peer.Peer] struct {
	closer.Monitor

	r *Resolver[T]

	mu      sync.Mutex
	ch      chan resolver.Update[T]
	initial bool
}

func (w *watcher[T]) Next(ctx context.Context, opts resolver.ResolveOptions) (resolver.Update[T], error) {
	w.mu.Lock()
	if !w.initial {
		w.initial = true
		ch, peers := w.r.subscribe()
		w.ch = ch
		w.mu.Unlock()

		if len(peers) > 0 {
			return resolver.Update[T]{Additions: peers}, nil
		}
	} else {
		w.mu.Unlock()
	}

	if opts.NoWait {
		select {
		case u := <-w.ch:
			return u, nil
		default:
			return resolver.Update[T]{}, resolver.ErrNoUpdates
		}
	}

	wctx := w.Context()
	rctx := w.r.Context()

	select {
	case u := <-w.ch:
		return u, nil
	case <-ctx.Done():
		return resolver.Update[T]{}, ctx.Err()
	case <-wctx.Done():
		return resolver.Update[T]{}, wctx.Err()
	case <-rctx.Done():
		w.Close()

		return resolver.Update[T]{}, rctx.Err()
	}
}

func (w *watcher[T]) Close() error {
	w.mu.Lock()
	ch := w.ch
	w.mu.Unlock()

	if ch != nil {
		w.r.unsubscribe(ch)
	}

	return w.Monitor.Close()
}
