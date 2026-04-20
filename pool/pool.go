package pool

import (
	"context"
	"io"

	"github.com/upfluence/pkg/v2/iopool"
)

type Pool[T any] interface {
	io.Closer

	Get(context.Context) (T, error)
	Put(T) error
	Discard(T) error
}

type transformPool[From, To any] struct {
	pool   Pool[From]
	fromTo func(From) To
	toFrom func(To) From
}

func NewTransformPool[From, To any](p Pool[From], fromTo func(From) To, toFrom func(To) From) Pool[To] {
	return &transformPool[From, To]{pool: p, fromTo: fromTo, toFrom: toFrom}
}

func (p *transformPool[From, To]) Get(ctx context.Context) (To, error) {
	var v, err = p.pool.Get(ctx)

	return p.fromTo(v), err
}

func (p *transformPool[From, To]) Put(v To) error     { return p.pool.Put(p.toFrom(v)) }
func (p *transformPool[From, To]) Discard(v To) error { return p.pool.Discard(p.toFrom(v)) }
func (p *transformPool[From, To]) Close() error       { return p.pool.Close() }

type entity[T comparable] struct {
	Value T
}

func (e entity[T]) Close() error               { return nil }
func (e entity[T]) Open(context.Context) error { return nil }
func (e entity[T]) IsOpen() bool               { return true }

type DowngradedPool[T comparable] struct {
	*iopool.Pool[entity[T]]
}

func NewPool[T comparable](newfn func() T, opts ...iopool.Option) *DowngradedPool[T] {
	return &DowngradedPool[T]{
		Pool: iopool.NewPool(
			func(context.Context) (entity[T], error) {
				return entity[T]{Value: newfn()}, nil
			},
			opts...,
		),
	}
}

func (p *DowngradedPool[T]) Get(ctx context.Context) (T, error) {
	var v, err = p.Pool.Get(ctx)

	return v.Value, err
}

func (p *DowngradedPool[T]) Put(v T) error     { return p.Pool.Put(entity[T]{Value: v}) }
func (p *DowngradedPool[T]) Discard(v T) error { return p.Pool.Discard(entity[T]{Value: v}) }
