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
