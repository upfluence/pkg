package cache

import (
	"io"
)

type Cache[K any, V any] interface {
	Get(K) (V, bool, error)
	Set(K, V) error
	Evict(K) error

	io.Closer
}
