package stringutil

// Deprecated: Prefer using generics.NullablePtr().
func NullablePtr(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
