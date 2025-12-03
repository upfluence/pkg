//go:build go1.18

package slices

// References returns slice with the value of the slice in params as
// pointers.
func References[T any](s []T) []*T {
	var r = make([]*T, len(s))

	for i, v := range s {
		v := v
		r[i] = &v
	}

	return r
}

func Indirect[T any](s []*T) []T {
	var r = make([]T, len(s))

	for i, v := range s {
		r[i] = *v
	}

	return r
}
