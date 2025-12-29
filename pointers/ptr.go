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

func Equal[T comparable](a, b *T) bool {
	if a == b {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if aa, ok := any(a).(interface{ Equal(b T) bool }); ok {
		return aa.Equal(*b)
	}

	return *a == *b
}
