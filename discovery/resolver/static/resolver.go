package static

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/upfluence/pkg/v2/closer"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/metadata"
)

type Builder[T peer.Peer] map[string][]T

func (b Builder[T]) Build(n string) resolver.Resolver[T] {
	return &Resolver[T]{Peers: b[n]}
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

	Peers []T
}

type Peer string

func (sp Peer) Addr() string                { return string(sp) }
func (sp Peer) Metadata() metadata.Metadata { return nil }

func NewResolverFromStrings(addrs []string) *Resolver[Peer] {
	return NewResolver(PeersFromStrings(addrs...))
}

func NewResolver[T peer.Peer](peers []T) *Resolver[T] {
	return &Resolver[T]{Peers: peers}
}

func (r *Resolver[T]) String() string {
	var addrs = make([]string, len(r.Peers))

	for i, peer := range r.Peers {
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

	r       *Resolver[T]
	initial int32
}

func (w *watcher[T]) Next(ctx context.Context, opts resolver.ResolveOptions) (resolver.Update[T], error) {
	ok := atomic.CompareAndSwapInt32(&w.initial, 0, 1)

	if opts.NoWait && (!ok || len(w.r.Peers) == 0) {
		return resolver.Update[T]{}, resolver.ErrNoUpdates
	}

	if ok && len(w.r.Peers) > 0 {
		return resolver.Update[T]{Additions: w.r.Peers}, nil
	}

	wctx := w.Context()
	rctx := w.r.Context()

	select {
	case <-ctx.Done():
		return resolver.Update[T]{}, ctx.Err()
	case <-wctx.Done():
		return resolver.Update[T]{}, wctx.Err()
	case <-rctx.Done():
		w.Close()
		return resolver.Update[T]{}, wctx.Err()
	}
}
