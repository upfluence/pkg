//go:build go1.18

package generics

import "github.com/upfluence/pkg/slices"

func Pointer[T comparable](v T) *T {
	return &v
}

func NullablePtr[T comparable](v T) *T {
	var zero T

	if v == zero {
		return nil
	}

	return &v
}

// ReferenceSlice returns slice with the value of the slice in params as
// pointers.
// Deprecated.
// Use slices.References instead.
func ReferenceSlice[T comparable](s []T) []*T {
	return slices.References(s)
}

// IndirectSlice
// Deprecated.
// Use slices.Indirect instead.
func IndirectSlice[T comparable](s []*T) []T {
	return slices.Indirect(s)
}
