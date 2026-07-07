package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/cache/policy/size"
)

type staticMultiFetcher[K comparable, V any] struct {
	fn func(context.Context, []K) (map[K]V, error)
}

func (sf *staticMultiFetcher[K, V]) Fetch(ctx context.Context, ks []K) (map[K]V, error) {
	return sf.fn(ctx, ks)
}

func TestWrapMultiFetcher(t *testing.T) {
	store := map[string]string{
		"foo":  "bar",
		"baz":  "qux",
		"quux": "corge",
	}

	evictC := make(chan string, 1)

	fetcher := WrapMultiFetcher[string, string](
		&staticMultiFetcher[string, string]{
			fn: func(_ context.Context, ks []string) (map[string]string, error) {
				res := make(map[string]string, len(ks))

				for _, k := range ks {
					if v, ok := store[k]; ok {
						res[k] = v
					}
				}

				return res, nil
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
	res, err := fetcher.Fetch(ctx, []string{"foo", "baz"})

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"foo": "bar", "baz": "qux"}, res)

	// Mutate the backing store; cached values must be returned.
	store["foo"] = "buz"
	store["baz"] = "changed"

	res, err = fetcher.Fetch(ctx, []string{"foo", "baz"})

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"foo": "bar", "baz": "qux"}, res)

	// Fetch a third key to evict "foo" (LRU cap 2, "baz" was just accessed).
	res, err = fetcher.Fetch(ctx, []string{"quux"})

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"quux": "corge"}, res)

	// Wait for the async eviction to complete.
	got := <-evictC

	assert.Equal(t, "foo", got)

	// "foo" is evicted; next fetch should hit the backing store.
	res, err = fetcher.Fetch(ctx, []string{"foo", "baz"})

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"foo": "buz", "baz": "qux"}, res)
}
