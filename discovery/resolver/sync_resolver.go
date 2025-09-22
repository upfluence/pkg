package resolver

import (
	"context"
	"sync"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/discovery/peer"
)

type SyncResolver[T peer.Peer] interface {
	ResolveSync(context.Context, string) ([]T, error)
	Close() error
}

func SyncResolverFromBuilder[T peer.Peer](b Builder[T], noWait bool) SyncResolver[T] {
	return &syncResolver[T]{
		builder: b,
		noWait:  noWait,
		lrs:     make(map[string]*localResolver[T]),
	}
}

type syncResolver[T peer.Peer] struct {
	builder Builder[T]
	noWait  bool

	mu  sync.Mutex
	lrs map[string]*localResolver[T]
}

func (sr *syncResolver[T]) ResolveSync(ctx context.Context, n string) ([]T, error) {
	sr.mu.Lock()

	lr, ok := sr.lrs[n]

	if !ok {
		lr = &localResolver[T]{readyc: make(chan struct{})}

		lr.p = &Puller[T]{
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

func (sr *syncResolver[T]) Close() error {
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

type localResolver[T peer.Peer] struct {
	p *Puller[T]

	readyOnce sync.Once
	readyc    chan struct{}

	mu sync.RWMutex
	ps map[string]T
}

func (lr *localResolver[T]) update(u Update[T]) {
	lr.mu.Lock()

	if lr.ps == nil {
		lr.ps = make(map[string]T)
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

func (lr *localResolver[T]) close() error {
	return errors.Combine(lr.p.Close())
}

func (lr *localResolver[T]) resolve(ctx context.Context) ([]T, error) {
	if !lr.p.IsOpen() {
		return nil, ErrClose
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-lr.readyc:
	}

	lr.mu.RLock()

	ps := make([]T, 0, len(lr.ps))

	for _, p := range lr.ps {
		ps = append(ps, p)
	}

	lr.mu.RUnlock()

	return ps, nil
}
