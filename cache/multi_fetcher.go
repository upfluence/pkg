package cache

import (
	"context"

	"github.com/upfluence/errors"
	"golang.org/x/exp/maps"

	"github.com/upfluence/pkg/v2/syncutil"
)

type MultiFetcher[K comparable, V any] interface {
	Fetch(context.Context, []K) (map[K]V, error)
}

type multiFetcher[K comparable, V any] struct {
	next  MultiFetcher[K, V]
	cache Cache[K, V]
	sf    syncutil.KeyedSingleflight[K, V]
}

func WrapMultiFetcher[K comparable, V any](f MultiFetcher[K, V], c Cache[K, V]) MultiFetcher[K, V] {
	return &multiFetcher[K, V]{next: f, cache: c}
}

func (cf *multiFetcher[K, V]) Fetch(ctx context.Context, ks []K) (map[K]V, error) {
	res := make(map[K]V, len(ks))

	var misses []K

	for _, k := range ks {
		v, ok, err := cf.cache.Get(k)

		if err != nil {
			return nil, errors.Wrap(err, "cache get")
		}

		if ok {
			res[k] = v
		} else {
			misses = append(misses, k)
		}
	}

	if len(misses) == 0 {
		return res, nil
	}

	fetched, err := cf.sf.Do(ctx, misses, func(ctx context.Context, ks []K) (map[K]V, error) {
		return cf.fetch(ctx, ks)
	})

	if err != nil {
		return nil, errors.Wrap(err, "singleflight")
	}

	maps.Copy(res, fetched)

	return res, nil
}

func (cf *multiFetcher[K, V]) fetch(ctx context.Context, ks []K) (map[K]V, error) {
	res := make(map[K]V, len(ks))

	var misses []K

	for _, k := range ks {
		v, ok, err := cf.cache.Get(k)

		if err != nil {
			return nil, errors.Wrap(err, "cache get")
		}

		if ok {
			res[k] = v
		} else {
			misses = append(misses, k)
		}
	}

	if len(misses) == 0 {
		return res, nil
	}

	fetched, err := cf.next.Fetch(ctx, misses)

	if err != nil {
		return nil, errors.Wrap(err, "fetch")
	}

	for k, v := range fetched {
		cf.cache.Set(k, v) //nolint:errcheck
		res[k] = v
	}

	return res, nil
}
