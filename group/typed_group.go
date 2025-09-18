package group

import (
	"context"
	"sync"
)

type TypedRunner[T any] func(context.Context) (func(*T), error)

type TypedGroup[T any] struct {
	Group Group
	Value T

	mu sync.Mutex
}

func (tg *TypedGroup[T]) Do(fn TypedRunner[T]) {
	tg.Group.Do(func(ctx context.Context) error {
		fn, err := fn(ctx)

		if err != nil {
			return err
		}

		tg.mu.Lock()
		fn(&tg.Value)
		tg.mu.Unlock()

		return nil
	})
}

func (tg *TypedGroup[T]) Wait() (T, error) {
	return tg.Value, tg.Group.Wait()
}
