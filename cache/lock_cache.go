package cache

import "sync"

type lockCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

func newLockCache[K comparable, V any]() *lockCache[K, V] {
	return &lockCache[K, V]{items: make(map[K]V)}
}

func (lc *lockCache[K, V]) get(res map[K]V, ks []K) {
	lc.mu.RLock()

	for _, k := range ks {
		if v, ok := lc.items[k]; ok {
			res[k] = v
		}
	}

	lc.mu.RUnlock()
}

func (lc *lockCache[K, V]) Get(k K) (V, bool, error) {
	lc.mu.RLock()
	v, ok := lc.items[k]
	lc.mu.RUnlock()

	return v, ok, nil
}

func (lc *lockCache[K, V]) Set(k K, v V) error {
	lc.mu.Lock()
	lc.items[k] = v
	lc.mu.Unlock()

	return nil
}

func (lc *lockCache[K, V]) Evict(k K) error {
	lc.mu.Lock()
	delete(lc.items, k)
	lc.mu.Unlock()

	return nil
}

func (lc *lockCache[K, V]) Close() error { return nil }
