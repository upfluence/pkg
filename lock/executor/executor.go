package executor

import (
	"context"
	"time"

	"github.com/upfluence/pkg/lock"
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

func WithNoWait(w bool) ExecutorOption {
	return func(eos *executorOptions) { eos.noWait = w }
}

type Executor struct {
	tw *TaskWrapper
	t  Task
}

func NewExecutor(l lock.Lock, t Task, opts ...ExecutorOption) *Executor {
	return &Executor{t: t, tw: NewTaskWrapper(l, opts...)}
}

func (e *Executor) Execute(ctx context.Context) error {
	return e.tw.Execute(ctx, e.t)
}
