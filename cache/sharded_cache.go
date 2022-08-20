package cache

import (
	"github.com/upfluence/errors"
	"github.com/upfluence/pkg/maputil"
	"github.com/upfluence/pkg/sliceutil"
	"golang.org/x/exp/constraints"
)

const (
	defaultSharding = 256

	offset64 uint64 = 14695981039346656037
	prime64  uint64 = 1099511628211
)

func fnv64a(s string) uint64 {
	var h = offset64

	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}

	return h
}

type shardedCache[K comparable, V any] struct {
	cs []*lockCache[K, V]

	kfn  func(K) uint64
	size uint64

	kp sliceutil.Pool[K]
	mp maputil.Pool[*lockCache[K, V], []K]
}

func NewStringCache[V any]() Cache[string, V] {
	return NewCache[string, V](fnv64a)
}

func NewIntegerCache[K constraints.Integer, V any]() Cache[K, V] {
	return NewCache[K, V](func(k K) uint64 { return uint64(k) })
}

func NewCache[K comparable, V any](kfn func(K) uint64) Cache[K, V] {
	return newShardedCache[K, V](defaultSharding, kfn)
}

func newShardedCache[K comparable, V any](size int, kfn func(K) uint64) Cache[K, V] {
	cs := make([]*lockCache[K, V], size)

	for i := 0; i < size; i++ {
		cs[i] = newLockCache[K, V]()
	}

	return &shardedCache[K, V]{cs: cs, kfn: kfn, size: uint64(size)}
}

func (sc *shardedCache[K, V]) shard(k K) *lockCache[K, V] {
	return sc.cs[sc.kfn(k)%sc.size]
}

func (sc *shardedCache[K, V]) Get(k K) (V, bool, error) {
	return sc.shard(k).Get(k)
}

func (sc *shardedCache[K, V]) Set(k K, v V) error {
	return sc.shard(k).Set(k, v)
}

func (sc *shardedCache[K, V]) Evict(k K) error {
	return sc.shard(k).Evict(k)
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
