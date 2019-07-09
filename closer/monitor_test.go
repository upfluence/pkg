package closer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmpty(t *testing.T) {
	m := NewMonior()

	assert.Nil(t, m.Shutdown(context.Background()))
	assert.Nil(t, m.Close())
}

func TestShutdownWait(t *testing.T) {
	m := NewMonior()
	closec := make(chan struct{})

	m.Run(func(context.Context) { <-closec })

	time.AfterFunc(10*time.Millisecond, func() { close(closec) })

	assert.Nil(t, m.Shutdown(context.Background()))
	assert.Nil(t, m.Close())
}

func TestShutdownTimeout(t *testing.T) {
	m := NewMonior()
	closec := make(chan struct{})

	m.Run(func(context.Context) { <-closec })

	assert.True(t, m.IsOpen())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	assert.Equal(t, context.DeadlineExceeded, m.Shutdown(ctx))
	assert.False(t, m.IsOpen())
	assert.Equal(t, Closing, m.s)

	close(closec)
	assert.Nil(t, m.Close())
	assert.False(t, m.IsOpen())
	assert.Equal(t, Closed, m.s)
}
