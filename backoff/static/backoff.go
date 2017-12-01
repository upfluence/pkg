package static

import (
	"time"

	"github.com/upfluence/pkg/backoff"
)

const Forever = -1

type Backoff struct {
	retries    int
	retryDelay time.Duration
}

func NewBackoff(retries int, delay time.Duration) *Backoff {
	return &Backoff{retries: retries, retryDelay: delay}
}

func NewInfiniteBackoff(delay time.Duration) *Backoff {
	return &Backoff{retries: Forever, retryDelay: delay}
}

func (b *Backoff) Backoff(n int) (time.Duration, error) {
	if b.retries == Forever || n < b.retries {
		return b.retryDelay, nil
	}

	return backoff.Cancelled, nil
}
