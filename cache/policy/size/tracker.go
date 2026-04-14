package size

import (
	"sync"

	"github.com/upfluence/pkg/v2/cache/policy"
)

// backend is the pure eviction-data-structure contract.
//
// Insert adds k and returns:
//   - evicted: the key that was displaced (zero value when ok is false)
//   - ok:      true if a key was displaced
//   - handle:  an opaque value identifying k's slot in the structure
//
// Remove deletes the slot identified by handle.
// Get records a hit for k, given its handle (e.g. move-to-back for LRU).
type backend[K comparable, V any] interface {
	insert(K) (K, bool, V)
	remove(V)
	get(K, V)
}

// tracker is a policy.Tracker[K] built on top of a backend.
// It owns the key→handle map and the evict callback.
type tracker[K comparable, V any] struct {
	mu sync.Mutex

	evict func(K)

	b  backend[K, V]
	ks map[K]V
	fn func(K)
}

func newPolicy[K comparable, V any](b backend[K, V], size int) policy.EvictionPolicy[K] {
	return policy.Wrap(func(evict func(K)) policy.Tracker[K] {
		t := &tracker[K, V]{
			b:     b,
			ks:    make(map[K]V, size),
			evict: evict,
		}
		t.fn = t.move
		return t
	})
}

func (t *tracker[K, V]) Op(k K, op policy.OpType) error {
	t.mu.Lock()

	var (
		evicted K
		ok      bool
	)

	switch op {
	case policy.Set:
		evicted, ok = t.insert(k)
	case policy.Get:
		t.fn(k)
	case policy.Evict:
		t.remove(k)
	}

	t.mu.Unlock()

	if ok {
		t.evict(evicted)
	}

	return nil
}

func (t *tracker[K, V]) Close() error { return nil }

// move moves k to the MRU position. Intended to be assigned to fn.
func (t *tracker[K, V]) move(k K) {
	n, ok := t.ks[k]
	if !ok {
		return
	}

	t.b.get(k, n)
}

// insert adds k. Must be called with t.mu held.
func (t *tracker[K, V]) insert(k K) (K, bool) {
	if _, exists := t.ks[k]; exists {
		t.fn(k)

		var zero K

		return zero, false
	}

	evicted, ok, n := t.b.insert(k)

	if ok {
		delete(t.ks, evicted)
	}

	t.ks[k] = n

	return evicted, ok
}

// remove deletes k. Must be called with t.mu held.
func (t *tracker[K, V]) remove(k K) {
	n, ok := t.ks[k]

	if !ok {
		return
	}

	t.b.remove(n)
	delete(t.ks, k)
}
