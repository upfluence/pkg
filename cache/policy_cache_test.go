package cache

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/v2/cache/policy/size"
)

func TestPolicyCache(t *testing.T) {
	var (
		done sync.WaitGroup
		once sync.Once
	)

	done.Add(1)

	c := WithEvictionPolicy[string](
		NewStringCache[string](),
		size.NewLRUPolicy[string](1),
		WithEvictionErrorHandler[string, string](func(_ string, err error) {
			t.Errorf("unexpected eviction error: %v", err)
		}),
		// Use a test hook via the error handler option — instead we need
		// a post-eviction notification.  Inject via a wrapper cache.
	)

	// Wrap the cache to intercept the eviction notification.
	// Since WithEvictionErrorHandler only fires on error, we instead use a
	// wrapping approach: override the cache with a spy.
	notifyC := make(chan struct{}, 1)

	c = WithEvictionPolicy[string](
		&evictSpy[string, string]{
			Cache: NewStringCache[string](),
			onEvict: func() {
				once.Do(func() {
					notifyC <- struct{}{}
				})
			},
		},
		size.NewLRUPolicy[string](1),
	)
	_ = done // silence unused warning; replaced by notifyC

	c.Set("foo", "bar")
	c.Set("bar", "buz")

	<-notifyC

	_, ok, err := c.Get("foo")

	assert.False(t, ok)
	assert.Nil(t, err)

	assert.Nil(t, c.Close())
}

// evictSpy wraps a Cache and calls onEvict whenever Evict is called.
type evictSpy[K comparable, V any] struct {
	Cache[K, V]
	onEvict func()
}

func (s *evictSpy[K, V]) Evict(k K) error {
	err := s.Cache.Evict(k)
	s.onEvict()
	return err
}
