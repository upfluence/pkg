package policy

import (
	"errors"
	"io"
)

type OpType uint8

const (
	Get OpType = iota
	Set
	Evict
)

type EvictionPolicy interface {
	C() <-chan string

	Op(string, OpType) error

	io.Closer
}

var ErrClosed = errors.New("cache/policy: eviction policy closed")
