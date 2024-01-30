//go:build go1.18

// Package generics
// Deprecated
package generics

import "github.com/upfluence/pkg/slices"

// Pointer pointers
//
// Deprecated: Use pointers.Ptr.
func Pointer[T comparable](v T) *T {
	return &v
}

// NullablePtr return nil if the value is zero
//
// Deprecated: Use pointers.NullablePtr.
func NullablePtr[T comparable](v T) *T {
	var zero T

	if v == zero {
		return nil
	}

	return &v
}

// ReferenceSlice returns slice with the value of the slice in params as
// pointers.
//
// Deprecated: use slices.References instead.
func ReferenceSlice[T comparable](s []T) []*T {
	return slices.References(s)
}

// IndirectSlice
//
// Deprecated: use slices.Indirect instead.
func IndirectSlice[T comparable](s []*T) []T {
	return slices.Indirect(s)
}
