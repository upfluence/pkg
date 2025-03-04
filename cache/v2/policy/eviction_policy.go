package policy

import (
	"io"

	"github.com/upfluence/errors"
)

type OpType uint8

const (
	Get OpType = iota
	Set
	Evict
)

type EvictionPolicy[K comparable] interface {
	C() <-chan K

	Op(K, OpType) error

	io.Closer
}

var ErrClosed = errors.New("cache/policy: eviction policy closed")
