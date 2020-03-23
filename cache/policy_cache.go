package cache

import (
	"sync"

	"github.com/upfluence/pkg/cache/policy"
)

func nop() {}

var testHookEviction = nop

type policyCache struct {
	c  Cache
	ep policy.EvictionPolicy

	wg sync.WaitGroup
}

func WithEvictionPolicy(c Cache, ep policy.EvictionPolicy) Cache {
	return newPolicyCache(c, ep)
}

func newPolicyCache(c Cache, ep policy.EvictionPolicy) *policyCache {
	pc := policyCache{c: c, ep: ep}

	pc.wg.Add(1)

	go pc.watch()

	return &pc
}

func (pc *policyCache) watch() {
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

func (pc *policyCache) Get(k string) (interface{}, bool, error) {
	v, ok, err := pc.c.Get(k)

	if err != nil {
		return v, ok, err
	}

	return v, ok, pc.ep.Op(k, policy.Get)
}

func (pc *policyCache) Set(k string, v interface{}) error {
	if err := pc.c.Set(k, v); err != nil {
		return err
	}

	return pc.ep.Op(k, policy.Set)
}

func (pc *policyCache) Evict(k string) error {
	if err := pc.c.Evict(k); err != nil {
		return err
	}

	return pc.ep.Op(k, policy.Evict)
}

func (pc *policyCache) Close() error {
	err := pc.ep.Close()

	pc.wg.Wait()
	return err
}
