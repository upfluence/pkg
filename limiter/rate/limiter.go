package rate

import (
	"context"
	"time"

	"golang.org/x/time/rate"

	"github.com/upfluence/pkg/limiter"
)

var noopDone = func() {}

type Config struct {
	Baseline int
	Period   time.Duration
	Burst    int
}

type Limiter struct {
	l *rate.Limiter
}

func NewLimiter(c Config) *Limiter {
	burst := c.Burst

	if burst == 0 {
		burst = c.Baseline
	}

	return &Limiter{
		l: rate.NewLimiter(
			rate.Limit(float64(c.Baseline)/float64(c.Period.Seconds())),
			burst,
		),
	}
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
