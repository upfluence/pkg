package cache

import (
	"github.com/upfluence/pkg/cache/policy"
	"github.com/upfluence/pkg/cache/v2"
)

type Cache cache.Cache[string, any]

func NewCache() Cache {
	return cache.NewCache[string, any]()
}

func WithEvictionPolicy(c Cache, ep policy.EvictionPolicy) Cache {
	return cache.WithEvictionPolicy(c, ep)
}
