package pool

import (
	"context"

	"github.com/upfluence/errors"
)

type ExecutorOption func(*executorOptions)

type executorOptions struct {
	shouldDiscard    func(error) bool
	getWrapError     func(error) error
	putWrapError     func(error) error
	discardWrapError func(error) error
}

func WithShouldDiscard(fn func(error) bool) ExecutorOption {
	return func(o *executorOptions) { o.shouldDiscard = fn }
}

func WithGetWrapError(fn func(error) error) ExecutorOption {
	return func(o *executorOptions) { o.getWrapError = fn }
}

func WithCheckinWrapError(fn func(error) error) ExecutorOption {
	return func(o *executorOptions) {
		o.putWrapError = fn
		o.discardWrapError = fn
	}
}

type Executor[T comparable] struct {
	executorOptions

	pool Pool[T]
}

func NewExecutor[T comparable](p Pool[T], opts ...ExecutorOption) *Executor[T] {
	var options = executorOptions{
		shouldDiscard:    func(err error) bool { return err != nil },
		getWrapError:     func(err error) error { return err },
		putWrapError:     func(err error) error { return err },
		discardWrapError: func(err error) error { return err },
	}

	for _, opt := range opts {
		opt(&options)
	}

	return &Executor[T]{pool: p, executorOptions: options}
}

func Execute[T comparable](ctx context.Context, p Pool[T], fn func(context.Context, T) error, opts ...ExecutorOption) error {
	return NewExecutor(p, opts...).Execute(ctx, fn)
}

func (e *Executor[T]) Execute(ctx context.Context, fn func(context.Context, T) error) error {
	var v, err = e.pool.Get(ctx)

	if err != nil {
		return e.getWrapError(err)
	}

	err = fn(ctx, v)

	if e.shouldDiscard(err) {
		return errors.Combine(err, e.discardWrapError(e.pool.Discard(v)))
	}

	return errors.Combine(err, e.putWrapError(e.pool.Put(v)))
}
