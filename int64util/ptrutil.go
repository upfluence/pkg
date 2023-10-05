package int64util

// Deprecated: Use generics.NullablePtr instead.
func NullablePtr(i int64) *int64 {
	if i == 0 {
		return nil
	}

	return &i
}
