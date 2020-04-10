package cache

import "io"

type Cache interface {
	Get(string) (interface{}, bool, error)
	Set(string, interface{}) error
	Evict(string) error

	io.Closer
}
