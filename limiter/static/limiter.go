package static

import (
	"context"

	"github.com/upfluence/pkg/limiter"
)

var (
	noopDone = func() {}

	OpenLimiter   = openLimiter{}
	ClosedLimiter = closedLimiter{}
)

type openLimiter struct{}

func (openLimiter) Allow(context.Context, limiter.AllowOptions) (limiter.DoneFunc, error) {
	return noopDone, nil
}

func (openLimiter) String() string { return "limiter/static/open" }

type closedLimiter struct{}

func (closedLimiter) Allow(ctx context.Context, opts limiter.AllowOptions) (limiter.DoneFunc, error) {
	if opts.NoWait {
		return nil, limiter.ErrLimited
	}

	<-ctx.Done()
	return nil, ctx.Err()
}

func (closedLimiter) String() string { return "limiter/static/open" }
