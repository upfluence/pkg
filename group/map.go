package group

import (
	"context"
	"sync"
)

type MapRunner[K comparable, V any] func(context.Context, K) (V, error)

func ExecuteMap[K comparable, V any](g Group, ks []K, fn MapRunner[K, V]) (map[K]V, error) {
	var (
		mu sync.Mutex

		res = make(map[K]V, len(ks))
	)

	for _, k := range ks {
		k := k

		g.Do(func(ctx context.Context) error {
			var v, err = fn(ctx, k)

			if err != nil {
				return err
			}

			mu.Lock()
			res[k] = v
			mu.Unlock()

			return nil
		})
	}

	return res, g.Wait()
}
