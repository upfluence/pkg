package int64util

// Deprecated: Prefer using generics.NullablePtr().
func NullablePtr(i int64) *int64 {
	if i == 0 {
		return nil
	}

	return &i
}
