//go:build go1.18

package generics

// Unique creates a map with keys being the values of a slice
//
// Deprecated: use slices.Unique.
func Unique[T comparable](s []T) map[T]struct{} {
	var m = make(map[T]struct{}, len(s))

	for _, v := range s {
		m[v] = struct{}{}
	}

	return m
}

// UniquePtr creates a map with keys being the de-referenced value of a pointer
// contained in a slice. Nil ptr are ignored.
//
// Deprecated: use slices.UniquePtr.
func UniquePtr[T comparable](s []*T) map[T]struct{} {
	var m = make(map[T]struct{}, len(s))

	for _, v := range s {
		if v == nil {
			continue
		}

		m[*v] = struct{}{}
	}

	return m
}
