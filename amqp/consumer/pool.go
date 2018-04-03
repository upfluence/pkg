package consumer

import (
	"context"

	stdpool "github.com/upfluence/pkg/pool"
)

type Pool interface {
	Get(context.Context) (Consumer, error)
	Put(Consumer) error
	Discard(Consumer) error
}

type pool struct {
	pool stdpool.Pool
	opts []Option
}

func NewPool(f stdpool.PoolFactory, opts ...Option) Pool {
	var p = &pool{opts: opts}

	p.pool = f.GetPool(p.factory)

	return p
}

func (p *pool) factory(ctx context.Context) (interface{}, error) {
	var c = NewConsumer(p.opts...)

	return c, c.Open(ctx)
}

func (p *pool) Get(ctx context.Context) (Consumer, error) {
	var c, err = p.pool.Get(ctx)

	if err != nil {
		return nil, err
	}

	return c.(Consumer), nil
}

func (p *pool) Put(c Consumer) error {
	return p.pool.Put(c)
}

func (p *pool) Discard(c Consumer) error {
	c.Close()
	return p.pool.Discard(c)
}
