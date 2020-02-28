package semaphore

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/limiter"
)

func TestLimiterNoWait(t *testing.T) {
	l := NewLimiter(3)
	opts := limiter.AllowOptions{NoWait: true, N: 2}
	ctx := context.Background()

	fn, err := l.Allow(ctx, opts)
	assert.Nil(t, err)

	fn2, err := l.Allow(ctx, opts)
	assert.Nil(t, fn2)
	assert.Equal(t, limiter.ErrLimited, err)

	assert.Equal(t, 1, l.remaining)
	fn()
	assert.Equal(t, 3, l.remaining)

	fn3, err := l.Allow(ctx, opts)
	assert.Nil(t, err)
	fn3()
}

func TestLimiteWaitCanceled(t *testing.T) {
	var (
		wg sync.WaitGroup

		l           = NewLimiter(1)
		opts        = limiter.AllowOptions{N: 2}
		ctx, cancel = context.WithCancel(context.Background())
	)

	wg.Add(1)

	go func() {
		fn, err := l.Allow(ctx, opts)
		assert.Equal(t, err, context.Canceled)
		assert.Nil(t, fn)
		wg.Done()
	}()

	cancel()

	wg.Wait()
}

func TestLimiteWaitPreCanceled(t *testing.T) {
	var (
		l           = NewLimiter(1)
		opts        = limiter.AllowOptions{N: 2}
		ctx, cancel = context.WithCancel(context.Background())
	)

	cancel()
	fn, err := l.Allow(ctx, opts)
	assert.Equal(t, err, context.Canceled)
	assert.Nil(t, fn)
}

func TestLimiteWaitSuccess(t *testing.T) {
	var (
		wg sync.WaitGroup

		l    = NewLimiter(3)
		opts = limiter.AllowOptions{N: 2}
		ctx  = context.Background()
	)

	fn, err := l.Allow(ctx, opts)
	assert.Equal(t, err, nil)

	wg.Add(1)

	go func() {
		fn2, err := l.Allow(ctx, opts)
		assert.Equal(t, err, nil)
		fn2()
		wg.Done()
	}()

	time.Sleep(time.Millisecond)
	fn()
	wg.Wait()
}
