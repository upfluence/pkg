package stringutil

// Deprecated: Use pointers.NullablePtr.
func NullablePtr(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
