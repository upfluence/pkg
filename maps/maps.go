package maps

type KV[K comparable, V any] struct {
	key K
	val V
}

func ToSlice[M ~map[K]V, K comparable, V any](m M) []KV[K, V] {
	ss := make([]KV[K, V], 0, len(m))

	for k, v := range m {
		ss = append(ss, KV[K, V]{k, v})
	}

	return ss
}
