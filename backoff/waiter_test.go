package backoff

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaiterCanceled(t *testing.T) {
	err := NewWaiter(
		LimitStrategy(
			StrategyFn(func(int) time.Duration { return time.Second }),
			0,
		),
	).Wait(context.Background())

	assert.Equal(t, ErrCanceled, err)
}

func TestWaiterContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewWaiter(
		StrategyFn(func(int) time.Duration { return time.Second }),
	).Wait(ctx)

	assert.Equal(t, context.Canceled, err)
}
