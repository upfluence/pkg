package multi

import (
	"context"
	"slices"
	"sync"

	"github.com/upfluence/pkg/v2/limiter"
)

// Limiter admits a call only when all its limiters do. With blocking calls it
// holds what it already reserved while waiting on the next limiter, callers
// sharing limiters must list them in the same order to avoid deadlocks.
type Limiter struct {
	rs []limiter.Reserver
}

func NewLimiter(rs ...limiter.Reserver) *Limiter { return &Limiter{rs: rs} }

func (l *Limiter) Allow(ctx context.Context, opts limiter.AllowOptions) (limiter.DoneFunc, error) {
	return limiter.WrapReserver(l).Allow(ctx, opts) //nolint:wrapcheck
}

func (l *Limiter) Reserve(ctx context.Context, opts limiter.AllowOptions) (limiter.DoneFunc, limiter.CancelFunc, error) {
	var (
		dones   = make([]limiter.DoneFunc, 0, len(l.rs))
		cancels = make([]limiter.CancelFunc, 0, len(l.rs))
	)

	for _, r := range l.rs {
		done, cancel, err := r.Reserve(ctx, opts)

		if err != nil {
			cancelAll(cancels)

			return nil, nil, err //nolint:wrapcheck
		}

		dones = append(dones, done)
		cancels = append(cancels, cancel)
	}

	var once sync.Once

	return func() {
			once.Do(func() {
				for _, done := range dones {
					done()
				}
			})
		}, func() {
			once.Do(func() { cancelAll(cancels) })
		}, nil
}

func cancelAll(cancels []limiter.CancelFunc) {
	for _, cancel := range slices.Backward(cancels) {
		cancel()
	}
}
