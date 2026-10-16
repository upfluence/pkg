package rate

import (
	"context"
	"fmt"
	"time"

	"github.com/upfluence/errors"
	"golang.org/x/time/rate"

	"github.com/upfluence/pkg/v2/limiter"
)

var noopDone = func() {}

type Config struct {
	Baseline int
	Period   time.Duration
	Burst    int
}

func (c Config) burst() int {
	if c.Burst == 0 {
		return c.Baseline
	}

	return c.Burst
}

func (c Config) limit() rate.Limit {
	return rate.Limit(float64(c.Baseline) / float64(c.Period.Seconds()))
}

type Limiter struct {
	l *rate.Limiter
}

func NewLimiter(c Config) *Limiter {
	return &Limiter{
		l: rate.NewLimiter(c.limit(), c.burst()),
	}
}

func (l *Limiter) String() string {
	return fmt.Sprintf("limiter/rate: [limit: %v, burst: %v, tokens: %d]", l.l.Limit(), l.l.Burst(), int(l.l.Tokens()))
}

func (l *Limiter) Update(c Config) {
	l.l.SetBurst(c.burst())
	l.l.SetLimit(c.limit())
}

func (l *Limiter) Allow(ctx context.Context, opts limiter.AllowOptions) (limiter.DoneFunc, error) {
	if opts.NoWait {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if l.l.AllowN(time.Now(), opts.N) {
			return noopDone, nil
		}

		return nil, limiter.ErrLimited
	}

	if err := l.l.WaitN(ctx, opts.N); err != nil {
		return nil, err
	}

	return noopDone, nil
}

// Reserve takes the tokens like Allow, the cancel func gives them back.
//
// The underlying limiter can not refund tokens already due, so the cancel
// func reserves a negative number of tokens, which adds them back to the
// bucket, the burst caps it.
func (l *Limiter) Reserve(ctx context.Context, opts limiter.AllowOptions) (limiter.DoneFunc, limiter.CancelFunc, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err() //nolint:wrapcheck
	default:
	}

	refund := func() { l.l.ReserveN(time.Now(), -opts.N) }
	now := time.Now()

	if opts.NoWait {
		if !l.l.AllowN(now, opts.N) {
			return nil, nil, limiter.ErrLimited
		}

		return noopDone, refund, nil
	}

	r := l.l.ReserveN(now, opts.N)

	if !r.OK() {
		return nil, nil, errors.Newf("limiter/rate: n=%d exceeds the burst %d", opts.N, l.l.Burst())
	}

	delay := r.DelayFrom(now)

	if d, ok := ctx.Deadline(); ok && d.Sub(now) < delay {
		refund()

		return nil, nil, context.DeadlineExceeded
	}

	if delay == 0 {
		return noopDone, refund, nil
	}

	t := time.NewTimer(delay)

	defer t.Stop()

	select {
	case <-t.C:
		return noopDone, refund, nil
	case <-ctx.Done():
		refund()

		return nil, nil, ctx.Err() //nolint:wrapcheck
	}
}
