//go:build go1.18

package generics

func Pointer[T comparable](v T) *T {
	return &v
}

// ReferenceSlice returns slice with the value of the slice in params as
// pointers.
func ReferenceSlice[T comparable](s []T) []*T {
	var r = make([]*T, len(s))

	for i, v := range s {
		v := v
		r[i] = &v
	}

	return r
}
