package stringutil

// Deprecated: Use generics.NullablePtr instead.
func NullablePtr(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
