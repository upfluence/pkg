package group

import (
	"context"
	"sync"

	"github.com/upfluence/errors"
)

type waitGroup struct {
	ctx context.Context
	fn  context.CancelFunc

	mu   sync.Mutex
	errs []error

	wg sync.WaitGroup
}

// WaitGroup creates a Group that waits for all goroutines to complete and
// collects all errors. Unlike ErrorGroup, it does not cancel remaining work
// when an error occurs - all goroutines run to completion.
//
// Wait returns a wrapped error containing all errors encountered, or nil if
// all goroutines completed successfully.
//
// This is useful when you want to gather all errors from concurrent operations
// rather than stopping at the first error.
//
// Example:
//
//	g := group.WaitGroup(ctx)
//
//	// Validate multiple items, collect all validation errors
//	for _, item := range items {
//		g.Do(func(ctx context.Context) error {
//			if err := validateItem(item); err != nil {
//				return fmt.Errorf("item %s invalid: %w", item.ID, err)
//			}
//			return nil
//		})
//	}
//
//	// Returns all errors encountered
//	if err := g.Wait(); err != nil {
//		return fmt.Errorf("validation failed: %w", err)
//	}
func WaitGroup(ctx context.Context) Group {
	var cctx, fn = context.WithCancel(ctx)

	return &waitGroup{ctx: cctx, fn: fn}
}

func (wg *waitGroup) Do(fn Runner) {
	select {
	case <-wg.ctx.Done():
		return
	default:
	}

	wg.wg.Add(1)

	go func() {
		defer wg.wg.Done()

		if err := fn(wg.ctx); err != nil {
			wg.mu.Lock()
			wg.errs = append(wg.errs, err)
			wg.mu.Unlock()
		}
	}()
}

func (wg *waitGroup) Wait() error {
	wg.wg.Wait()
	wg.fn()

	return errors.WrapErrors(wg.errs)
}
