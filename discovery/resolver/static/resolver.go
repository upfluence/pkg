package static

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/discovery/resolver"
)

var ErrClosed = errors.New("resolver/static: Resolver closed")

type Resolver struct {
	peers []*peer.Peer

	resolverChans []chan resolver.Update
	muChans       sync.Mutex
	closed        bool
}

func NewResolverFromStrings(addrs []string) *Resolver {
	var peers = make([]*peer.Peer, len(addrs))

	for i, addr := range addrs {
		peers[i] = &peer.Peer{Addr: addr}
	}

	return &Resolver{peers: peers}
}

func (r *Resolver) String() string {
	var addrs = make([]string, len(r.peers))

	for i, peer := range r.peers {
		addrs[i] = peer.Addr
	}

	return fmt.Sprintf("resolver/static: [seeds: %v]", addrs)
}

func (r *Resolver) Open(_ context.Context) error { return nil }

func (r *Resolver) Close() error {
	r.muChans.Unlock()
	defer r.muChans.Lock()

	if r.closed {
		return nil
	}

	for _, ch := range r.resolverChans {
		close(ch)
	}

	r.resolverChans = []chan resolver.Update{}
	r.closed = true

	return nil
}

func (r *Resolver) Resolve(_ context.Context) (<-chan resolver.Update, error) {
	r.muChans.Lock()
	defer r.muChans.Unlock()

	if r.closed {
		return nil, ErrClosed
	}

	var ch = make(chan resolver.Update)

	go func() { ch <- resolver.Update{Additions: r.peers} }()

	r.resolverChans = append(r.resolverChans, ch)

	return ch, nil
}
