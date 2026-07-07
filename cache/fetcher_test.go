package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/cache/policy/size"
)

type staticFetcher[K, V any] struct {
	fn func(context.Context, K) (V, error)
}

func (sf *staticFetcher[K, V]) Fetch(ctx context.Context, k K) (V, error) {
	return sf.fn(ctx, k)
}

func TestWrapFetcher(t *testing.T) {
	store := map[string]string{
		"foo":  "bar",
		"baz":  "qux",
		"quux": "corge",
	}

	evictC := make(chan string, 1)

	fetcher := WrapFetcher[string, string](
		&staticFetcher[string, string]{
			fn: func(_ context.Context, k string) (string, error) {
				return store[k], nil
			},
		},
		WithEvictionPolicy[string, string](
			&evictSpy[string, string]{
				Cache:   NewStringCache[string](),
				onEvict: func(k string) { evictC <- k },
			},
			size.NewLRUPolicy[string](2),
		),
	)

	ctx := context.Background()

	// First fetch populates the cache.
	v, err := fetcher.Fetch(ctx, "foo")

	require.NoError(t, err)
	assert.Equal(t, "bar", v)

	// Mutate the backing store; cached value must be returned.
	store["foo"] = "buz"

	v, err = fetcher.Fetch(ctx, "foo")

	require.NoError(t, err)
	assert.Equal(t, "bar", v)

	// Fill the cache with two other keys so "foo" gets evicted (LRU cap 2).
	_, err = fetcher.Fetch(ctx, "baz")

	require.NoError(t, err)

	_, err = fetcher.Fetch(ctx, "quux")

	require.NoError(t, err)

	// Wait for the async eviction of "foo" to complete.
	got := <-evictC

	assert.Equal(t, "foo", got)

	// Now "foo" is evicted; the next fetch hits the backing store.
	v, err = fetcher.Fetch(ctx, "foo")

	require.NoError(t, err)
	assert.Equal(t, "buz", v)
}
