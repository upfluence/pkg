package bounded

import (
	"context"

	"github.com/upfluence/pkg/iopool"
	"github.com/upfluence/pkg/pool"
	"github.com/upfluence/stats"
)

type PoolFactory struct {
	size int
}

func NewPoolFactory(size int) *PoolFactory {
	return &PoolFactory{size: size}
}

func (f *PoolFactory) GetPool(factory pool.Factory) pool.Pool {
	return f.GetIntrospectablePool(factory)
}

func (f *PoolFactory) GetIntrospectablePool(factory pool.Factory) pool.IntrospectablePool {
	return NewPool(f.size, factory)
}

type entityWrapper struct {
	v interface{}
}

func (e entityWrapper) Close() error               { return nil }
func (e entityWrapper) Open(context.Context) error { return nil }
func (e entityWrapper) IsOpen() bool               { return true }

func NewPool(limit int, factory pool.Factory) pool.IntrospectablePool {
	c := stats.NewStaticCollector()

	return &poolWrapper{
		Pool: iopool.NewPool(
			func(ctx context.Context) (iopool.Entity, error) {
				v, err := factory(ctx)
				if err != nil {
					return nil, err
				}

				return entityWrapper{v}, nil
			},
			iopool.WithSize(limit),
			iopool.WithMaxIdle(limit),
			iopool.WithScope(stats.RootScope(c)),
		),
		c: c,
	}
}

type poolWrapper struct {
	*iopool.Pool

	c *stats.StaticCollector
}

func (p *poolWrapper) Get(ctx context.Context) (interface{}, error) {
	v, err := p.Pool.Get(ctx)

	if err != nil {
		return nil, err
	}

	return v.(*entityWrapper).v, nil
}

func (p *poolWrapper) Put(v interface{}) error {
	return p.Pool.Put(entityWrapper{v: v})
}

func (p *poolWrapper) Discard(v interface{}) error {
	return p.Pool.Discard(entityWrapper{v: v})
}

func (p *poolWrapper) GetStats() (int, int) {
	snap := p.c.Get()

	var idle, checkout int

	for _, is := range snap.Gauges {
		switch is.Name {
		case "idle":
			idle = int(is.Value)
		case "checkout":
			checkout = int(is.Value)
		}
	}

	return idle, checkout
}
