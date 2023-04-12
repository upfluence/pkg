package executor

import (
	"context"
	"errors"
	"sync"

	"github.com/upfluence/pkg/lock"
	"github.com/upfluence/pkg/log"
)

type TaskWrapper struct {
	l lock.Lock

	executorOptions
}

func NewTaskWrapper(l lock.Lock, opts ...ExecutorOption) *TaskWrapper {
	var tw = TaskWrapper{l: l, executorOptions: defaultExecutorOptions}

	for _, opt := range opts {
		opt(&tw.executorOptions)
	}

	return &tw
}

func (tw *TaskWrapper) Execute(ctx context.Context, t Task) error {
	var (
		wg sync.WaitGroup

		log     = log.WithContext(ctx).WithField(log.Field("lock", tw.l))
		le, err = tw.l.Acquire(ctx, tw.acquireOptions())
	)

	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)

	wg.Add(2)

	go func() {
		select {
		case <-le.Done():
			cancel()
		case <-cctx.Done():
		}

		wg.Done()
	}()

	go func() {
		t := tw.clock.Timer(tw.renewInterval)

		for {
			select {
			case <-cctx.Done():
				t.Stop()
				wg.Done()

				return
			case <-t.C():
				kerr := le.KeepAlive(ctx, tw.deadline())

				switch {
				case kerr == nil || errors.Is(kerr, context.Canceled):
				case errors.Is(kerr, lock.ErrLeaseNotFound):
					select {
					case <-cctx.Done():
					default:
						log.WithError(kerr).Error("lost lease")
					}

					cancel()
				default:
					log.WithError(kerr).Error("cant renew lock")
				}
			}

			t.Reset(tw.renewInterval)
		}
	}()

	err = t(cctx)

	cancel()
	wg.Wait()

	if rerr := le.Release(
		ctx,
	); rerr != nil && !errors.Is(rerr, context.Canceled) &&
		!errors.Is(rerr, lock.ErrLeaseNotFound) {
		log.WithError(rerr).Error("cant release lock")
	}

	return err
}
