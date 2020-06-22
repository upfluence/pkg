package static

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/upfluence/pkg/closer"
	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/discovery/resolver"
	"github.com/upfluence/pkg/metadata"
)

type Builder map[string][]peer.Peer

func (b Builder) Build(n string) resolver.Resolver {
	return &Resolver{Peers: b[n]}
}

func PeersFromStrings(addrs ...string) []peer.Peer {
	var peers = make([]peer.Peer, len(addrs))

	for i, addr := range addrs {
		peers[i] = staticPeer(addr)
	}

	return peers
}

type Resolver struct {
	closer.Monitor

	Peers []peer.Peer
}

type staticPeer string

func (sp staticPeer) Addr() string                { return string(sp) }
func (sp staticPeer) Metadata() metadata.Metadata { return nil }

func NewResolverFromStrings(addrs []string) *Resolver {
	return NewResolver(PeersFromStrings(addrs...))
}

func NewResolver(peers []peer.Peer) *Resolver {
	return &Resolver{Peers: peers}
}

func (r *Resolver) String() string {
	var addrs = make([]string, len(r.Peers))

	for i, peer := range r.Peers {
		addrs[i] = peer.Addr()
	}

	return fmt.Sprintf("resolver/static: [seeds: %v]", addrs)
}

func (r *Resolver) Build(string) resolver.Resolver { return r }

func (r *Resolver) Open(_ context.Context) error { return nil }

func (r *Resolver) Resolve() resolver.Watcher {
	return &watcher{r: r}
}

type watcher struct {
	closer.Monitor

	r       *Resolver
	initial int32
}

func (w *watcher) Next(ctx context.Context, opts resolver.ResolveOptions) (resolver.Update, error) {
	ok := atomic.CompareAndSwapInt32(&w.initial, 0, 1)

	if opts.NoWait && (!ok || len(w.r.Peers) == 0) {
		return resolver.Update{}, resolver.ErrNoUpdates
	}

	if ok && len(w.r.Peers) > 0 {
		return resolver.Update{Additions: w.r.Peers}, nil
	}

	wctx := w.Context()
	rctx := w.r.Context()

	select {
	case <-ctx.Done():
		return resolver.Update{}, ctx.Err()
	case <-wctx.Done():
		return resolver.Update{}, wctx.Err()
	case <-rctx.Done():
		w.Close()
		return resolver.Update{}, wctx.Err()
	}
}
