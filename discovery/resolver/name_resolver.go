package resolver

import (
	"context"
	"sync"

	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/multierror"
)

type NameResolver struct {
	Builder Builder

	mu  sync.Mutex
	lrs map[string]*localResolver
}

func (nr *NameResolver) Resolve(ctx context.Context, n string) ([]peer.Peer, error) {
	nr.mu.Lock()

	if nr.lrs == nil {
		nr.lrs = make(map[string]*localResolver)
	}

	lr, ok := nr.lrs[n]

	if !ok {
		lr = &localResolver{readyc: make(chan struct{})}
		lr.p = &Puller{Resolver: nr.Builder.Build(n), UpdateFunc: lr.update}

		if err := lr.p.Open(ctx); err != nil {
			nr.mu.Unlock()
			close(lr.readyc)
			return nil, err
		}

		nr.lrs[n] = lr
	}

	nr.mu.Unlock()

	return lr.resolve(ctx)
}

func (nr *NameResolver) Close() error {
	var errs []error

	nr.mu.Lock()

	for _, lr := range nr.lrs {
		if err := lr.close(); err != nil {
			errs = append(errs)
		}
	}

	nr.lrs = nil
	nr.mu.Unlock()

	return multierror.Wrap(errs)
}

type localResolver struct {
	p *Puller

	readyOnce sync.Once
	readyc    chan struct{}

	mu sync.RWMutex
	ps map[string]peer.Peer
}

func (lr *localResolver) update(u Update) {
	lr.mu.Lock()

	if lr.ps == nil {
		lr.ps = make(map[string]peer.Peer)
	}

	for _, p := range u.Deletions {
		addr := p.Addr()

		if _, ok := lr.ps[addr]; ok {
			delete(lr.ps, addr)
		}
	}

	for _, p := range u.Additions {
		lr.ps[p.Addr()] = p
	}

	lr.mu.Unlock()

	if len(lr.ps) > 0 {
		lr.readyOnce.Do(func() { close(lr.readyc) })
	}
}

func (lr *localResolver) close() error {
	return multierror.Combine(lr.p.Close())
}

func (lr *localResolver) resolve(ctx context.Context) ([]peer.Peer, error) {
	if !lr.p.IsOpen() {
		return nil, ErrClose
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-lr.readyc:
	}

	lr.mu.RLock()

	ps := make([]peer.Peer, 0, len(lr.ps))

	for _, p := range lr.ps {
		ps = append(ps, p)
	}

	lr.mu.RUnlock()

	return ps, nil
}
