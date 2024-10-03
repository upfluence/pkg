package slices

import "golang.org/x/exp/constraints"

func ReduceWith[S ~[]E, E any, A any](s S, acc A, fn func(A, E) A) A {
	for _, e := range s {
		acc = fn(acc, e)
	}

	return acc
}

func Reduce[S ~[]E, E any, A any](s S, fn func(A, E) A) A {
	var acc A

	return ReduceWith(s, acc, fn)
}

type sumable interface {
	constraints.Integer | constraints.Float | ~string
}

func Sum[T sumable](i, j T) T {
	return i + j
}
