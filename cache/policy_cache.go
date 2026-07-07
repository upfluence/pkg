package cache

import (
	"sync"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/cache/policy"
)

type policyCache[K any, V any] struct {
	c  Cache[K, V]
	ep policy.EvictionPolicy[K]

	onEvict func(K, error)

	wg sync.WaitGroup
}

// PolicyCacheOption configures a policyCache.
type PolicyCacheOption[K any, V any] func(*policyCache[K, V])

// WithEvictionErrorHandler sets a callback that is invoked whenever the
// policy-triggered eviction of a key fails.  The default is to discard the
// error silently.
func WithEvictionErrorHandler[K any, V any](fn func(K, error)) PolicyCacheOption[K, V] {
	return func(pc *policyCache[K, V]) {
		pc.onEvict = fn
	}
}

func WithEvictionPolicy[K any, V any](c Cache[K, V], ep policy.EvictionPolicy[K], opts ...PolicyCacheOption[K, V]) Cache[K, V] {
	return newPolicyCache(c, ep, opts...)
}

func newPolicyCache[K any, V any](c Cache[K, V], ep policy.EvictionPolicy[K], opts ...PolicyCacheOption[K, V]) *policyCache[K, V] {
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
		return v, false, errors.Wrap(err, "cache get")
	}

	if ok {
		if err := pc.ep.Op(k, policy.Get); err != nil {
			return v, false, errors.Wrap(err, "policy op")
		}
	}

	return v, ok, nil
}

func (pc *policyCache[K, V]) Set(k K, v V) error {
	if err := pc.c.Set(k, v); err != nil {
		return errors.Wrap(err, "cache set")
	}

	return errors.Wrap(pc.ep.Op(k, policy.Set), "policy op")
}

func (pc *policyCache[K, V]) Evict(k K) error {
	if err := pc.c.Evict(k); err != nil {
		return errors.Wrap(err, "cache evict")
	}

	return errors.Wrap(pc.ep.Op(k, policy.Evict), "policy op")
}

// Close closes the eviction policy (which stops the background pump and closes
// the channel), waits for the watch goroutine to drain, then closes the
// underlying cache.
func (pc *policyCache[K, V]) Close() error {
	err := pc.ep.Close()
	pc.wg.Wait()

	if cerr := pc.c.Close(); cerr != nil {
		err = errors.Combine(err, cerr)
	}

	return err
}
