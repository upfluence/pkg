package static

import (
	"time"

	"github.com/upfluence/pkg/v2/backoff"
)

func NewBackoff(ts int, delay time.Duration) backoff.Strategy {
	return backoff.LimitStrategy(NewInfiniteBackoff(delay), ts)
}

func NewInfiniteBackoff(delay time.Duration) backoff.Strategy {
	return backoff.StrategyFn(func(int) time.Duration { return delay })
}
