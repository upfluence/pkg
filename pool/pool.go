package pool

import (
	"context"

	"github.com/upfluence/pkg/v2/iopool"
)

type entity[T comparable] struct {
	Value T
}

func (e entity[T]) Close() error               { return nil }
func (e entity[T]) Open(context.Context) error { return nil }
func (e entity[T]) IsOpen() bool               { return true }

type Pool[T comparable] struct {
	*iopool.Pool[entity[T]]
}

func NewPool[T comparable](newfn func() T, opts ...iopool.Option) *Pool[T] {
	return &Pool[T]{
		Pool: iopool.NewPool(
			func(context.Context) (entity[T], error) {
				return entity[T]{Value: newfn()}, nil
			},
			opts...,
		),
	}
}

func (p *Pool[T]) Get(ctx context.Context) (T, error) {
	var v, err = p.Pool.Get(ctx)

	return v.Value, err
}

func (p *Pool[T]) Put(v T) error     { return p.Pool.Put(entity[T]{Value: v}) }
func (p *Pool[T]) Discard(v T) error { return p.Pool.Discard(entity[T]{Value: v}) }
