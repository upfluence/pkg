package executor

import (
	"context"
	"sync"
	"time"

	"github.com/upfluence/pkg/lock"
	"github.com/upfluence/pkg/log"
	"github.com/upfluence/pkg/timeutil"
)

var defaultExecutorOptions = executorOptions{
	lockDuration:  30 * time.Second,
	renewInterval: 10 * time.Second,
	clock:         timeutil.Background(),
}

type executorOptions struct {
	clock timeutil.Clock

	lockDuration  time.Duration
	renewInterval time.Duration

	noWait bool
	op     lock.OpType
}

func (eos executorOptions) deadline() time.Time {
	return eos.clock.Now().Add(eos.lockDuration)
}

func (eos executorOptions) acquireOptions() lock.AcquireOptions {
	return lock.AcquireOptions{
		NoWait:   eos.noWait,
		Deadline: eos.deadline(),
		Op:       eos.op,
	}
}

type Task func(context.Context) error
type ExecutorOption func(*executorOptions)

func WithLockDuration(d time.Duration) ExecutorOption {
	return func(eos *executorOptions) { eos.lockDuration = d }
}

func WithRenewInterval(d time.Duration) ExecutorOption {
	return func(eos *executorOptions) { eos.renewInterval = d }
}

type Executor struct {
	l lock.Lock
	t Task

	executorOptions
}

func NewExecutor(l lock.Lock, t Task, opts ...ExecutorOption) *Executor {
	var e = Executor{l: l, t: t, executorOptions: defaultExecutorOptions}

	for _, opt := range opts {
		opt(&e.executorOptions)
	}

	return &e
}

func (e *Executor) Execute(ctx context.Context) error {
	var (
		wg sync.WaitGroup

		le, err      = e.l.Acquire(ctx, e.acquireOptions())
		cctx, cancel = context.WithCancel(ctx)
	)

	if err != nil {
		cancel()
		return err
	}

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
		t := e.clock.Timer(e.renewInterval)

		for {
			select {
			case <-cctx.Done():
				t.Stop()
				wg.Done()
				return
			case <-t.C():
				if kerr := le.KeepAlive(ctx, e.deadline()); kerr != nil {
					log.WithError(kerr).Error("cant renew lock")
				}
			}

			t.Reset(e.renewInterval)
		}
	}()

	err = e.t(cctx)

	cancel()

	if rerr := le.Release(ctx); rerr != nil {
		log.WithError(rerr).Error("cant release lock")
	}

	wg.Wait()

	return err
}
