package static

import (
	"context"

	"github.com/upfluence/pkg/v2/limiter"
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

func (openLimiter) Reserve(context.Context, limiter.AllowOptions) (limiter.DoneFunc, limiter.CancelFunc, error) {
	return noopDone, noopDone, nil
}

func (openLimiter) String() string { return "limiter/static/open" }

type closedLimiter struct{}

func (closedLimiter) Allow(ctx context.Context, opts limiter.AllowOptions) (limiter.DoneFunc, error) {
	if opts.NoWait {
		return nil, limiter.ErrLimited
	}

	<-ctx.Done()

	return nil, ctx.Err() //nolint:wrapcheck
}

func (cl closedLimiter) Reserve(ctx context.Context, opts limiter.AllowOptions) (limiter.DoneFunc, limiter.CancelFunc, error) {
	_, err := cl.Allow(ctx, opts)

	return nil, nil, err
}

func (closedLimiter) String() string { return "limiter/static/open" }
