package group

import (
	"context"
	"sync"

	"github.com/upfluence/errors"
)

// MapRunner is a function type that processes a key of type K and returns
// a value of type V. This is used with ExecuteMap to concurrently process
// a slice of keys and build a result map.
//
// Example:
//
//	var runner group.MapRunner[int, string] = func(ctx context.Context, id int) (string, error) {
//		user, err := fetchUser(ctx, id)
//		if err != nil {
//			return "", err
//		}
//		return user.Name, nil
//	}
type MapRunner[K comparable, V any] func(context.Context, K) (V, error)

// ExecuteMap concurrently processes a slice of keys using the provided
// MapRunner and returns a map of results. Each key is processed in a
// separate goroutine managed by the provided Group.
//
// The function blocks until all keys are processed and returns the result
// map and any errors from the Group. If any runner returns an error, the
// behavior depends on the Group implementation (e.g., ErrorGroup will
// stop on first error, WaitGroup will collect all errors).
//
// Example:
//
//	userIDs := []int{1, 2, 3, 4, 5}
//
//	users, err := group.ExecuteMap(
//		group.ErrorGroup(ctx),
//		userIDs,
//		func(ctx context.Context, id int) (*User, error) {
//			return fetchUser(ctx, id)
//		},
//	)
//	if err != nil {
//		return fmt.Errorf("failed to fetch users: %w", err)
//	}
//
//	// users is map[int]*User with results for each ID
//	for id, user := range users {
//		fmt.Printf("User %d: %s\n", id, user.Name)
//	}
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
