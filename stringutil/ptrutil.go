package stringutil

func NullablePtr(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
