package cache

import (
	"context"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/syncutil"
)

type Fetcher[K, V any] interface {
	Fetch(context.Context, K) (V, error)
}

type fetcher[K comparable, V any] struct {
	next  Fetcher[K, V]
	cache Cache[K, V]
	sf    syncutil.KeyedSingleflight[K, V]
}

func WrapFetcher[K comparable, V any](f Fetcher[K, V], c Cache[K, V]) Fetcher[K, V] {
	return &fetcher[K, V]{next: f, cache: c}
}

func (cf *fetcher[K, V]) Fetch(ctx context.Context, k K) (V, error) {
	v, ok, err := cf.cache.Get(k)

	if err != nil {
		return v, errors.Wrap(err, "cache get")
	}

	if ok {
		return v, nil
	}

	_, v, err = syncutil.DoOne(ctx, &cf.sf, k, func(ctx context.Context) (V, error) {
		return cf.fetch(ctx, k)
	})

	if err != nil {
		return v, errors.Wrap(err, "singleflight")
	}

	return v, nil
}

func (cf *fetcher[K, V]) fetch(ctx context.Context, k K) (V, error) {
	v, ok, err := cf.cache.Get(k)

	if err != nil {
		return v, errors.Wrap(err, "cache get")
	}

	if ok {
		return v, nil
	}

	v, err = cf.next.Fetch(ctx, k)

	if err != nil {
		return v, errors.Wrap(err, "fetch")
	}

	cf.cache.Set(k, v) //nolint:errcheck

	return v, nil
}
