package resolver

import (
	"context"
	"sync"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/discovery/peer"
)

type SyncResolver interface {
	ResolveSync(context.Context, string) ([]peer.Peer, error)
	Close() error
}

func SyncResolverFromBuilder(b Builder, noWait bool) SyncResolver {
	return &syncResolver{
		builder: b,
		noWait:  noWait,
		lrs:     make(map[string]*localResolver),
	}
}

type syncResolver struct {
	builder Builder
	noWait  bool

	mu  sync.Mutex
	lrs map[string]*localResolver
}

func (sr *syncResolver) ResolveSync(ctx context.Context, n string) ([]peer.Peer, error) {
	sr.mu.Lock()

	lr, ok := sr.lrs[n]

	if !ok {
		lr = &localResolver{readyc: make(chan struct{})}

		lr.p = &Puller{
			Resolver:   sr.builder.Build(n),
			UpdateFunc: lr.update,
			NoWait:     sr.noWait,
		}

		if err := lr.p.Open(ctx); err != nil {
			sr.mu.Unlock()
			close(lr.readyc)
			return nil, err
		}

		sr.lrs[n] = lr
	}

	sr.mu.Unlock()

	return lr.resolve(ctx)
}

func (sr *syncResolver) Close() error {
	var errs []error

	sr.mu.Lock()

	for _, lr := range sr.lrs {
		if err := lr.close(); err != nil {
			errs = append(errs)
		}
	}

	sr.lrs = nil
	sr.mu.Unlock()

	return errors.WrapErrors(errs)
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

	lr.readyOnce.Do(func() { close(lr.readyc) })
}

func (lr *localResolver) close() error {
	return errors.Combine(lr.p.Close())
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
