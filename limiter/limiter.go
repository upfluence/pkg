package limiter

import (
	"context"

	"github.com/upfluence/errors"
)

var ErrLimited = errors.New("limiter: Limited")

type DoneFunc func()

type AllowOptions struct {
	NoWait bool
	N      int
}

type Limiter interface {
	Allow(context.Context, AllowOptions) (DoneFunc, error)
}

type CancelFunc func()

// Reserver admits calls like a Limiter, and lets the caller undo the
// admission, so several limiters can be acquired all or none. A reservation
// ends with exactly one of the two functions: the DoneFunc once the call is
// over, or the CancelFunc to undo it.
type Reserver interface {
	Reserve(context.Context, AllowOptions) (DoneFunc, CancelFunc, error)
}

// WrapReserver turns a Reserver into a Limiter, the admission is never undone.
func WrapReserver(r Reserver) Limiter { return reserverLimiter{r} }

type reserverLimiter struct {
	Reserver
}

func (rl reserverLimiter) Allow(ctx context.Context, opts AllowOptions) (DoneFunc, error) {
	done, _, err := rl.Reserve(ctx, opts)

	return done, err
}
