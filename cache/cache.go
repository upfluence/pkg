package cache

import (
	"io"
)

type Cache[K comparable, V any] interface {
	Get(K) (V, bool, error)
	Set(K, V) error
	Evict(K) error

	io.Closer
}
