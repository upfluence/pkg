package group

import (
	"context"
	"sync"
)

type errGroup struct {
	ctx    context.Context
	fn     context.CancelFunc
	err    error
	stopfn func(error) bool

	wg   sync.WaitGroup
	once sync.Once
}

// ExitGroup creates a Group that cancels all goroutines as soon as any
// goroutine completes (with or without error). This is useful when you want
// the first completion to trigger cancellation of all other work.
//
// The returned error from Wait will be the error from the first goroutine
// that completed, or nil if it completed successfully.
func ExitGroup(ctx context.Context) Group {
	return newErrGroup(ctx, func(err error) bool { return true })
}

// ErrorGroup creates a Group that cancels all goroutines as soon as any
// goroutine returns an error. If all goroutines complete successfully,
// Wait returns nil.
//
// This is the most commonly used error handling strategy - fail fast on
// the first error encountered.
func ErrorGroup(ctx context.Context) Group {
	return newErrGroup(ctx, func(err error) bool { return err != nil })
}

func newErrGroup(ctx context.Context, stopfn func(error) bool) Group {
	var cctx, fn = context.WithCancel(ctx)

	return &errGroup{ctx: cctx, fn: fn, stopfn: stopfn}
}

func (eg *errGroup) Do(fn Runner) {
	select {
	case <-eg.ctx.Done():
		return
	default:
	}

	eg.wg.Add(1)

	go func() {
		defer eg.wg.Done()

		if err := fn(eg.ctx); eg.stopfn(err) {
			eg.once.Do(func() {
				eg.fn()
				eg.err = err
			})
		}
	}()
}

func (eg *errGroup) Wait() error {
	eg.wg.Wait()
	eg.fn()

	return eg.err
}
