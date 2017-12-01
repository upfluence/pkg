package backoff

import "time"

const Cancelled time.Duration = -1

type Strategy interface {
	Backoff(int) (time.Duration, error)
}
