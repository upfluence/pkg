package cache

import (
	"github.com/upfluence/pkg/multierror"
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

type shardedCache struct {
	cs []Cache

	kfn  func(string) uint64
	size uint64
}

func NewCache() Cache {
	return newShardedCache(newLockCache, defaultSharding)
}

func newShardedCache(cfn func() Cache, size int) Cache {
	cs := make([]Cache, size)

	for i := 0; i < size; i++ {
		cs[i] = cfn()
	}

	return &shardedCache{cs: cs, kfn: fnv64a, size: uint64(size)}
}

func (sc *shardedCache) Get(k string) (interface{}, bool, error) {
	return sc.cs[sc.kfn(k)%sc.size].Get(k)
}

func (sc *shardedCache) Set(k string, v interface{}) error {
	return sc.cs[sc.kfn(k)%sc.size].Set(k, v)
}

func (sc *shardedCache) Evict(k string) error {
	return sc.cs[sc.kfn(k)%sc.size].Evict(k)
}

func (sc *shardedCache) Close() error {
	var errs []error

	for _, c := range sc.cs {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return multierror.Wrap(errs)
}
