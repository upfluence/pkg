package cache

import "sync"

type lockCache struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

func newLockCache() Cache {
	return &lockCache{items: make(map[string]interface{})}
}

func (lc *lockCache) Get(k string) (interface{}, bool, error) {
	lc.mu.RLock()
	v, ok := lc.items[k]
	lc.mu.RUnlock()

	return v, ok, nil
}

func (lc *lockCache) Set(k string, v interface{}) error {
	lc.mu.Lock()
	lc.items[k] = v
	lc.mu.Unlock()

	return nil
}

func (lc *lockCache) Evict(k string) error {
	lc.mu.Lock()
	delete(lc.items, k)
	lc.mu.Unlock()

	return nil
}

func (lc *lockCache) Close() error { return nil }
