package pool

import (
	"context"
	"io"
)

type Factory func(context.Context) (interface{}, error)

type Pool interface {
	io.Closer

	Get(context.Context) (interface{}, error)
	Put(interface{}) error
	Discard(interface{}) error
}

type PoolFactory interface {
	GetPool(Factory) Pool
}

type IntrospectablePool interface {
	Pool

	GetStats() (int, int)
}

type IntrospectablePoolFactory interface {
	GetIntrospectablePool(Factory) IntrospectablePool
}
