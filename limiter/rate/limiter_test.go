package rate

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/limiter"
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
