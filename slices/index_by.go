package slices

// IndexBy create a map indexed by the key derived from the value
func IndexBy[T any, K comparable](ss []T, fn func(T) K) map[K]T {
	var res = make(map[K]T, len(ss))

	for _, t := range ss {
		res[fn(t)] = t
	}

	return res
}
