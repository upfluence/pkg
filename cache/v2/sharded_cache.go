package cache

import (
	"hash/maphash"

	"github.com/upfluence/errors"
)

const defaultSharding = 256

type shardedCache[K comparable, V any] struct {
	cs []Cache[K, V]

	kfn  func(K) uint64
	size uint64
}

func NewCache[K comparable, V any]() Cache[K, V] {
	return newShardedCache[K, V](newLockCache, defaultSharding)
}

func newShardedCache[K comparable, V any](cfn func() Cache[K, V], size int) Cache[K, V] {
	cs := make([]Cache[K, V], size)

	for i := 0; i < size; i++ {
		cs[i] = cfn()
	}

	seed := maphash.MakeSeed()

	return &shardedCache[K, V]{cs: cs, kfn: func(k K) uint64 {
		return maphash.Comparable(seed, k)
	}, size: uint64(size)}
}

func (sc *shardedCache[K, V]) Get(k K) (V, bool, error) {
	return sc.cs[sc.kfn(k)%sc.size].Get(k)
}

func (sc *shardedCache[K, V]) Set(k K, v V) error {
	return sc.cs[sc.kfn(k)%sc.size].Set(k, v)
}

func (sc *shardedCache[K, V]) Evict(k K) error {
	return sc.cs[sc.kfn(k)%sc.size].Evict(k)
}

func (sc *shardedCache[K, V]) Close() error {
	var errs []error

	for _, c := range sc.cs {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.WrapErrors(errs)
}
