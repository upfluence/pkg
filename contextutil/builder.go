package contextutil

import (
	"context"
	"time"
)

type ContextBuilder func() (context.Context, context.CancelFunc)

func Timeout(d time.Duration) ContextBuilder {
	return func() (context.Context, context.CancelFunc) {
		return context.WithTimeout(context.Background(), 5*time.Second)
	}
}
