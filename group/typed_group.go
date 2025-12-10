package group

import (
	"context"
	"sync"
)

// TypedRunner is a function type that executes work and returns a function
// to mutate a shared value of type T. The mutation function is called
// under a lock to ensure thread-safe updates.
//
// The pattern is: do work concurrently, then safely merge results into
// a shared accumulator.
type TypedRunner[T any] func(context.Context) (func(*T), error)

// TypedGroup manages concurrent execution of TypedRunner functions that
// collaboratively build a shared value of type T.
//
// Each runner executes concurrently, and upon successful completion, its
// mutation function is called under a lock to safely update the shared Value.
type TypedGroup[T any] struct {
	Group Group
	Value T

	mu sync.Mutex
}

// Do schedules a TypedRunner to execute concurrently. Upon successful
// completion, the runner's mutation function is called under a lock to
// safely update the shared Value.
//
// If the runner returns an error, the mutation function is not called.
//
// Example:
//
//	tg.Do(func(ctx context.Context) (func(*map[string]int), error) {
//		count, err := countItems(ctx)
//		if err != nil {
//			return nil, err
//		}
//		return func(m *map[string]int) {
//			(*m)["total"] += count
//		}, nil
//	})
func (tg *TypedGroup[T]) Do(runner TypedRunner[T]) {
	tg.Group.Do(func(ctx context.Context) error {
		fn, err := runner(ctx)

		if err != nil {
			return err
		}

		tg.mu.Lock()
		fn(&tg.Value)
		tg.mu.Unlock()

		return nil
	})
}

// Wait blocks until all scheduled runners complete and returns the final
// accumulated Value and any errors from the underlying Group.
func (tg *TypedGroup[T]) Wait() (T, error) {
	return tg.Value, tg.Group.Wait()
}
