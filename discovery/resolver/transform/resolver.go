package transform

import (
	"context"

	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

type transformResolver[S, T peer.Peer] struct {
	source    resolver.Resolver[S]
	transform func(S) T
}

func WrapResolver[S, T peer.Peer](r resolver.Resolver[S], fn func(S) T) resolver.Resolver[T] {
	return &transformResolver[S, T]{
		source:    r,
		transform: fn,
	}
}

func (r *transformResolver[S, T]) Open(ctx context.Context) error {
	return r.source.Open(ctx)
}

func (r *transformResolver[S, T]) Close() error {
	return r.source.Close()
}

func (r *transformResolver[S, T]) Resolve() resolver.Watcher[T] {
	return &watcher[S, T]{
		inner:     r.source.Resolve(),
		transform: r.transform,
	}
}

type watcher[S, T peer.Peer] struct {
	inner     resolver.Watcher[S]
	transform func(S) T
}

func (w *watcher[S, T]) Close() error {
	return w.inner.Close()
}

func (w *watcher[S, T]) Next(ctx context.Context, opts resolver.ResolveOptions) (resolver.Update[T], error) {
	u, err := w.inner.Next(ctx, opts)

	if err != nil {
		return resolver.Update[T]{}, err
	}

	var result resolver.Update[T]

	result.Additions = make([]T, len(u.Additions))

	for i, p := range u.Additions {
		result.Additions[i] = w.transform(p)
	}

	result.Deletions = make([]T, len(u.Deletions))

	for i, p := range u.Deletions {
		result.Deletions[i] = w.transform(p)
	}

	return result, nil
}
