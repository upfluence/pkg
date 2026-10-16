package static

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/limiter"
)

func TestReserve(t *testing.T) {
	done, cancel, err := OpenLimiter.Reserve(context.Background(), limiter.AllowOptions{})

	require.NoError(t, err)
	assert.NotNil(t, done)
	assert.NotNil(t, cancel)

	done, cancel, err = ClosedLimiter.Reserve(context.Background(), limiter.AllowOptions{NoWait: true})

	require.ErrorIs(t, err, limiter.ErrLimited)
	assert.Nil(t, done)
	assert.Nil(t, cancel)
}
