package local

import (
	"context"
	"sync"
	"time"

	"github.com/upfluence/pkg/lock"
)

type LockManager struct {
	init sync.Once

	mu  sync.Mutex
	les map[string]*leaseEntry
}

func (lm *LockManager) Lock(n string) lock.Lock {
	return &lockImpl{n: n, lm: lm}
}

type leaseEntry struct {
	sync.Mutex

	l  *lease
	ch chan struct{}
}

type lockImpl struct {
	n  string
	lm *LockManager
}

func (l *lockImpl) String() string { return l.n }

func (l *lockImpl) Acquire(ctx context.Context, opts lock.AcquireOptions) (lock.Lease, error) {
	return l.lm.acquire(ctx, l.n, opts)
}

func (lm *LockManager) acquire(ctx context.Context, n string, opts lock.AcquireOptions) (lock.Lease, error) {
	lm.init.Do(func() { lm.les = make(map[string]*leaseEntry) })

	d := time.Until(opts.Deadline)

	if d <= 0 {
		return nil, lock.ErrPastTime
	}

	lm.mu.Lock()
	le, ok := lm.les[n]

	if !ok {
		l := lease{n: n, lm: lm, d: opts.Deadline}

		l.Context, l.cancel = context.WithCancel(context.Background())
		l.Lock()
		l.t = time.AfterFunc(d, func() { l.Release(nil) })
		l.Unlock()

		lm.les[n] = &leaseEntry{ch: make(chan struct{}), l: &l}
		lm.mu.Unlock()

		return &l, nil
	}

	if opts.NoWait {
		lm.mu.Unlock()
		return nil, lock.ErrAlreadyAcquired
	}

	ch := le.ch
	lm.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		return lm.acquire(ctx, n, opts)
	}
}

func (lm *LockManager) release(l *lease) {
	lm.mu.Lock()
	le, ok := lm.les[l.n]

	if !ok {
		lm.mu.Unlock()
		return
	}

	delete(lm.les, l.n)
	close(le.ch)
	lm.mu.Unlock()
}

type lease struct {
	context.Context
	cancel context.CancelFunc

	n string

	lm *LockManager

	sync.Mutex
	t *time.Timer
	d time.Time
}

func (l *lease) Release(context.Context) error {
	if err := l.Err(); err != nil {
		return err
	}

	l.cancel()

	l.Lock()
	l.t.Stop()
	l.d = time.Time{}
	l.Unlock()

	l.lm.release(l)

	return nil
}

func (l *lease) KeepAlive(_ context.Context, deadline time.Time) error {
	d := time.Until(deadline)

	if d <= 0 {
		return lock.ErrPastTime
	}

	if err := l.Err(); err != nil {
		return err
	}

	l.Lock()
	l.t.Reset(d)
	l.d = deadline
	l.Unlock()

	return nil
}

func (l *lease) Deadline() time.Time {
	l.Lock()
	d := l.d
	l.Unlock()

	return d
}
