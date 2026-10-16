package rate

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upfluence/pkg/v2/limiter"
)

func TestString(t *testing.T) {
	l := NewLimiter(Config{Baseline: 2, Period: time.Second})
	done, err := l.Allow(context.Background(), limiter.AllowOptions{N: 1})

	defer done()

	assert.NoError(t, err)
	assert.Equal(t, "limiter/rate: [limit: 2, burst: 2, tokens: 1]", l.String())

	l.Update(Config{Baseline: 3, Period: time.Second})

	assert.Equal(t, "limiter/rate: [limit: 3, burst: 3, tokens: 1]", l.String())
}

func TestReserve(t *testing.T) {
	ctx := context.Background()
	l := NewLimiter(Config{Baseline: 1, Period: time.Hour, Burst: 2})

	_, cancel, err := l.Reserve(ctx, limiter.AllowOptions{NoWait: true, N: 2})

	require.NoError(t, err)

	_, _, err = l.Reserve(ctx, limiter.AllowOptions{NoWait: true, N: 1})

	require.ErrorIs(t, err, limiter.ErrLimited)

	// a failed NoWait reservation does not keep its tokens
	cancel()

	_, _, err = l.Reserve(ctx, limiter.AllowOptions{NoWait: true, N: 2})

	require.NoError(t, err)

	_, _, err = l.Reserve(ctx, limiter.AllowOptions{N: 3})

	require.Error(t, err)
	assert.NotErrorIs(t, err, limiter.ErrLimited)
}

func TestReserveWait(t *testing.T) {
	l := NewLimiter(Config{Baseline: 1, Period: 20 * time.Millisecond, Burst: 1})

	_, _, err := l.Reserve(context.Background(), limiter.AllowOptions{N: 1})

	require.NoError(t, err)

	start := time.Now()
	_, _, err = l.Reserve(context.Background(), limiter.AllowOptions{N: 1})

	require.NoError(t, err)
	assert.GreaterOrEqual(t, time.Since(start), 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)

	defer cancel()

	_, _, err = l.Reserve(ctx, limiter.AllowOptions{N: 1})

	require.ErrorIs(t, err, context.DeadlineExceeded)
}
