package backoff

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/upfluence/pkg/timeutil"
)

var ErrCanceled = errors.New("backoff: strategy canceled")

type Waiter struct {
	s Strategy
	c timeutil.Clock

	i int32
}

func NewWaiter(s Strategy) *Waiter {
	return &Waiter{s: s, c: timeutil.Background()}
}

func (w *Waiter) Reset() { atomic.StoreInt32(&w.i, 0) }

func (w *Waiter) Wait(ctx context.Context) error {
	i := atomic.AddInt32(&w.i, 1)

	d, err := w.s.Backoff(int(i))

	if err != nil {
		return err
	}

	if d == Canceled {
		return ErrCanceled
	}

	t := w.c.Timer(d)

	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C():
		return nil
	}
}
