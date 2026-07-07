package cache

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/v2/cache/policy/size"
)

func TestPolicyCache(t *testing.T) {
	var once sync.Once

	// Wrap the cache to intercept the eviction notification.
	notifyC := make(chan struct{}, 1)

	c := WithEvictionPolicy[string](
		&evictSpy[string, string]{
			Cache: NewStringCache[string](),
			onEvict: func(string) {
				once.Do(func() {
					notifyC <- struct{}{}
				})
			},
		},
		size.NewLRUPolicy[string](1),
		WithEvictionErrorHandler[string, string](func(_ string, err error) {
			t.Errorf("unexpected eviction error: %v", err)
		}),
	)

	c.Set("foo", "bar")
	c.Set("bar", "buz")

	<-notifyC

	_, ok, err := c.Get("foo")

	assert.False(t, ok)
	assert.NoError(t, err)

	assert.NoError(t, c.Close())
}

// evictSpy wraps a Cache and calls onEvict whenever Evict is called.
type evictSpy[K comparable, V any] struct {
	Cache[K, V]
	onEvict func(K)
}

func (s *evictSpy[K, V]) Evict(k K) error {
	err := s.Cache.Evict(k)
	s.onEvict(k)

	return err
}
