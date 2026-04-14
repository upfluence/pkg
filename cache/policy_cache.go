package cache

import (
	"sync"

	"github.com/upfluence/pkg/v2/cache/policy"
)

type policyCache[K comparable, V any] struct {
	c  Cache[K, V]
	ep policy.EvictionPolicy[K]

	onEvict func(K, error)

	wg sync.WaitGroup
}

// PolicyCacheOption configures a policyCache.
type PolicyCacheOption[K comparable, V any] func(*policyCache[K, V])

// WithEvictionErrorHandler sets a callback that is invoked whenever the
// policy-triggered eviction of a key fails.  The default is to discard the
// error silently.
func WithEvictionErrorHandler[K comparable, V any](fn func(K, error)) PolicyCacheOption[K, V] {
	return func(pc *policyCache[K, V]) {
		pc.onEvict = fn
	}
}

func WithEvictionPolicy[K comparable, V any](c Cache[K, V], ep policy.EvictionPolicy[K], opts ...PolicyCacheOption[K, V]) Cache[K, V] {
	return newPolicyCache(c, ep, opts...)
}

func newPolicyCache[K comparable, V any](c Cache[K, V], ep policy.EvictionPolicy[K], opts ...PolicyCacheOption[K, V]) *policyCache[K, V] {
	pc := policyCache[K, V]{
		c:       c,
		ep:      ep,
		onEvict: func(K, error) {},
	}

	for _, opt := range opts {
		opt(&pc)
	}

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

		if err := pc.c.Evict(k); err != nil {
			pc.onEvict(k, err)
		}
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

// Close closes the eviction policy (which stops the background pump and closes
// the channel), waits for the watch goroutine to drain, then closes the
// underlying cache.
func (pc *policyCache[K, V]) Close() error {
	err := pc.ep.Close()
	pc.wg.Wait()

	if cerr := pc.c.Close(); cerr != nil && err == nil {
		err = cerr
	}

	return err
}
