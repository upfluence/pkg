//go:build go1.20

package slices

// MapSlice transforms a slice of a given type to new slice of a different type
func MapSlice[T, K any](ss []T, fn func(v T) (K, error)) ([]K, error) {
	var res = make([]K, len(ss))

	for i, v := range ss {
		vv, err := fn(v)

		if err != nil {
			return nil, err
		}

		res[i] = vv
	}

	return res, nil
}
