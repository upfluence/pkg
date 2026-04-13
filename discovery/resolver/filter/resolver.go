package filter

import (
	"context"

	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

type filterResolver[T peer.Peer] struct {
	inner resolver.Resolver[T]
	allow func(T) bool
}

func WrapResolver[T peer.Peer](r resolver.Resolver[T], allow func(T) bool) resolver.Resolver[T] {
	return &filterResolver[T]{inner: r, allow: allow}
}

func (r *filterResolver[T]) Open(ctx context.Context) error {
	return r.inner.Open(ctx)
}

func (r *filterResolver[T]) Close() error {
	return r.inner.Close()
}

func (r *filterResolver[T]) Resolve() resolver.Watcher[T] {
	return &watcher[T]{inner: r.inner.Resolve(), allow: r.allow}
}

type watcher[T peer.Peer] struct {
	inner    resolver.Watcher[T]
	allow    func(T) bool
	admitted map[string]struct{}
}

func (w *watcher[T]) Close() error {
	return w.inner.Close()
}

func (w *watcher[T]) Next(ctx context.Context, opts resolver.ResolveOptions) (resolver.Update[T], error) {
	// When NoWait is requested we only attempt one read from the inner watcher.
	// If that read is empty (ErrNoUpdates) or entirely filtered out, we return
	// ErrNoUpdates immediately.  We must not keep consuming inner updates in a
	// loop under NoWait — doing so would silently drain the inner channel and
	// discard real updates the caller may want to inspect later.
	if opts.NoWait {
		u, err := w.inner.Next(ctx, opts)
		if err != nil {
			return resolver.Update[T]{}, err
		}

		filtered := w.filter(u)
		if len(filtered.Additions) == 0 && len(filtered.Deletions) == 0 {
			return resolver.Update[T]{}, resolver.ErrNoUpdates
		}

		return filtered, nil
	}

	// Blocking path: loop until we get at least one non-empty filtered update.
	for {
		u, err := w.inner.Next(ctx, opts)
		if err != nil {
			return resolver.Update[T]{}, err
		}

		filtered := w.filter(u)
		if len(filtered.Additions) == 0 && len(filtered.Deletions) == 0 {
			continue
		}

		return filtered, nil
	}
}

func (w *watcher[T]) filter(u resolver.Update[T]) resolver.Update[T] {
	var out resolver.Update[T]

	for _, p := range u.Additions {
		if w.allow(p) {
			if w.admitted == nil {
				w.admitted = make(map[string]struct{})
			}

			w.admitted[p.Addr()] = struct{}{}
			out.Additions = append(out.Additions, p)
		}
	}

	for _, p := range u.Deletions {
		if _, ok := w.admitted[p.Addr()]; ok {
			delete(w.admitted, p.Addr())
			out.Deletions = append(out.Deletions, p)
		}
	}

	return out
}
