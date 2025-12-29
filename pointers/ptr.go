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

func Eq[T comparable](a, b *T) bool {
	if a == b {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}
