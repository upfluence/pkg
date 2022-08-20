package cache

import (
	"sync"

	"github.com/upfluence/pkg/cache/policy"
)

func nop() {}

var testHookEviction = nop

type policyCache[K comparable, V any] struct {
	c  Cache[K, V]
	ep policy.EvictionPolicy[K]

	wg sync.WaitGroup
}

func WithEvictionPolicy[K comparable, V any](c Cache[K, V], ep policy.EvictionPolicy[K]) Cache[K, V] {
	return newPolicyCache(c, ep)
}

func newPolicyCache[K comparable, V any](c Cache[K, V], ep policy.EvictionPolicy[K]) *policyCache[K, V] {
	pc := policyCache[K, V]{c: c, ep: ep}

	pc.wg.Add(1)

	go pc.watch()

	return &pc
}

func (pc *policyCache[K, V]) watch() {
	defer pc.wg.Done()

	ch := pc.ep.C()

	for {
		k, ok := <-ch

		if !ok {
			return
		}

		pc.c.Evict(k)
		testHookEviction()
	}
}

func (pc *policyCache[K, V]) Get(k K) (V, bool, error) {
	v, ok, err := pc.c.Get(k)

	if err != nil {
		return v, false, err
	}

	if ok {
		if err := pc.ep.Op(k, policy.Get); err != nil {
			return v, false, err
		}
	}

	return v, ok, nil
}

func (pc *policyCache[K, V]) Set(k K, v V) error {
	if err := pc.c.Set(k, v); err != nil {
		return err
	}

	return pc.ep.Op(k, policy.Set)
}

func (pc *policyCache[K, V]) Evict(k K) error {
	if err := pc.c.Evict(k); err != nil {
		return err
	}

	return pc.ep.Op(k, policy.Evict)
}

func (pc *policyCache[K, V]) Close() error {
	err := pc.ep.Close()

	pc.wg.Wait()
	return err
}
