package stringcache

import "github.com/upfluence/errors"

var ErrNotFound = errors.New("stringcache: Key not found")

type Cache interface {
	Has(string) (bool, error)
	Add(string) error
	Delete(string) error
}
