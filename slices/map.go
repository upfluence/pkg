//go:build go1.21

package slices

import "context"

// MapWithContextError transforms a slice of a given type to a new slice of a
// different type using a context and can return an error.
func MapWithContextError[T, K any](ctx context.Context, ss []T, fn func(context.Context, T) (K, error)) ([]K, error) {
	var res = make([]K, len(ss))

	for i, v := range ss {
		vv, err := fn(ctx, v)

		if err != nil {
			return nil, err
		}

		res[i] = vv
	}

	return res, nil
}

// MapWithError transforms a slice of a given type to new slice of a different
// type, can return an error.
func MapWithError[T, K any](ss []T, fn func(T) (K, error)) ([]K, error) {
	return MapWithContextError(
		context.Background(),
		ss,
		func(ctx context.Context, t T) (K, error) {
			return fn(t)
		},
	)
}

// Map transforms a slice of a given type to new slice of a different type.
func Map[T, K any](ss []T, fn func(T) K) []K {
	r, _ := MapWithError(
		ss,
		func(t T) (K, error) {
			return fn(t), nil
		},
	)

	return r
}
