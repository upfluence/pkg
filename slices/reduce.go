package slices

// ReduceFrom will iterate over each element of slice and call reduceFn for each.
// The return value of each reduceFn is passed to the next call of reduceFn as acc, or
// in other term at each iteration on the slice.
func ReduceFrom[S ~[]E, E any, A any](slice S, acc A, reduceFn func(acc A, elem E) A) A {
	for _, e := range slice {
		acc = reduceFn(acc, e)
	}

	return acc
}

// Reduce does the same as ReduceFrom but use the zero value of A as
// the first value to be passed to reduceFn.
func Reduce[S ~[]E, E any, A any](slice S, reduceFn func(acc A, elem E) A) A {
	var acc A

	return ReduceFrom(slice, acc, reduceFn)
}
