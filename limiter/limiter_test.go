package limiter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeReserver struct {
	opts      AllowOptions
	done      int
	cancel    int
	reservErr error
}

func (fr *fakeReserver) Reserve(_ context.Context, opts AllowOptions) (DoneFunc, CancelFunc, error) {
	fr.opts = opts

	if fr.reservErr != nil {
		return nil, nil, fr.reservErr
	}

	return func() { fr.done++ }, func() { fr.cancel++ }, nil
}

func TestWrapReserver(t *testing.T) {
	fr := &fakeReserver{}
	l := WrapReserver(fr)

	done, err := l.Allow(context.Background(), AllowOptions{NoWait: true, N: 3})

	require.NoError(t, err)
	assert.Equal(t, AllowOptions{NoWait: true, N: 3}, fr.opts)

	done()

	assert.Equal(t, 1, fr.done)
	assert.Zero(t, fr.cancel)

	fr.reservErr = ErrLimited

	done, err = l.Allow(context.Background(), AllowOptions{})

	require.ErrorIs(t, err, ErrLimited)
	assert.Nil(t, done)
}
