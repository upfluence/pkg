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

func ExitGroup(ctx context.Context) Group {
	return newErrGroup(ctx, func(err error) bool { return true })
}

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
