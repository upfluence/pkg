package maps

import "golang.org/x/exp/maps"

// Keys is a dropin replacement for x/maps.Keys
func Keys[M ~map[K]T, K comparable, T any](m M) []K {
	return maps.Keys(m)
}

// Values is a dropin replacement for x/maps.Values
func Values[M ~map[K]T, K comparable, T any](m M) []T {
	return maps.Values(m)
}
