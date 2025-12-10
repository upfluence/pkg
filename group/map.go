package group

import (
	"context"
	"sync"

	"github.com/upfluence/errors"
)

// MapRunner is a function type that processes a key of type K and returns
// a value of type V. This is used with ExecuteMap to concurrently process
// a slice of keys and build a result map.
type MapRunner[K comparable, V any] func(context.Context, K) (V, error)

// ExecuteMap concurrently processes a slice of keys using the provided
// MapRunner and returns a map of results. Each key is processed in a
// separate goroutine managed by the provided Group.
//
// The function blocks until all keys are processed and returns the result
// map and any errors from the Group. If any runner returns an error, the
// behavior depends on the Group implementation (e.g., ErrorGroup will
// stop on first error, WaitGroup will collect all errors).
func ExecuteMap[K comparable, V any](g Group, ks []K, fn MapRunner[K, V]) (map[K]V, error) {
	var (
		mu sync.Mutex

		res = make(map[K]V, len(ks))
	)

	for _, k := range ks {
		g.Do(func(ctx context.Context) error {
			var v, err = fn(ctx, k)

			if err != nil {
				return errors.WithStack(err)
			}

			mu.Lock()
			res[k] = v
			mu.Unlock()

			return nil
		})
	}

	return res, g.Wait()
}
