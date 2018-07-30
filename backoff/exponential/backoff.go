package exponential

import (
	"time"

	"github.com/upfluence/pkg/backoff"
)

func NewDefaultBackoff(min, max time.Duration) backoff.Strategy {
	return NewBackoff(min, max, func(i int) int64 { return 1 << uint(i) })
}

func NewBackoff(min, max time.Duration, nextFn func(int) int64) backoff.Strategy {
	return backoff.StrategyFn(
		func(i int) time.Duration {
			var d = time.Duration(nextFn(i)) * min

			if d > max {
				return max
			}

			return d
		},
	)
}
