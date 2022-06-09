package generics

// Keys build a slice containing all the keys in a given map
func Keys[T comparable](m map[T]any) []T {
	ks := make([]T, 0, len(m))

	for k := range m {
		ks = append(ks, k)
	}

	return ks
}
