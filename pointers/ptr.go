package pointers

func Ptr[T comparable](v T) *T {
	return &v
}

func NullablePtr[T comparable](v T) *T {
	var zero T

	if v == zero {
		return nil
	}

	return &v
}

func NullIsZero[T comparable](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}
