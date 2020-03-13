package backoff

import (
	"math/rand"
	"time"
)

const (
	Canceled time.Duration = -1
	Forever                = -1

	Cancelled = Canceled
)

type Strategy interface {
	Backoff(int) (time.Duration, error)
}

func LimitStrategy(s Strategy, l int) Strategy {
	return &limitedStrategy{Strategy: s, limit: l}
}

type limitedStrategy struct {
	Strategy

	limit int
}

func (b *limitedStrategy) Backoff(n int) (time.Duration, error) {
	if l := b.limit; l == Forever || n < l {
		return b.Strategy.Backoff(n)
	}

	return Canceled, nil
}

type StrategyFn func(int) time.Duration

func (fn StrategyFn) Backoff(i int) (time.Duration, error) {
	return fn(i), nil
}

func JitterStrategy(s Strategy, j time.Duration) Strategy {
	return &jitterStrategy{
		Strategy: s,
		jitFn: func() time.Duration {
			return time.Duration(rand.Int63n(int64(j)))
		},
	}
}

type jitterStrategy struct {
	Strategy

	jitFn func() time.Duration
}

func (s *jitterStrategy) Backoff(i int) (time.Duration, error) {
	var d, err = s.Strategy.Backoff(i)

	if err != nil {
		return 0, err
	}

	return d + s.jitFn(), nil
}
