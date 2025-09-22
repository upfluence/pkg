package rate

import (
	"context"
	"fmt"
	"time"

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
