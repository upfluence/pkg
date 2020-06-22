package syncutil

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCondYield(t *testing.T) {
	var mu sync.Mutex

	cond := &Cond{Locker: &mu}

	go func() {
		time.Sleep(10 * time.Millisecond)

		cond.Signal()
	}()

	err := cond.Wait(context.Background(), func() bool { return true })

	assert.Nil(t, err)
}

func TestCondExpired(t *testing.T) {
	var mu sync.Mutex

	cond := &Cond{Locker: &mu}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)

		cancel()
	}()

	err := cond.Wait(ctx, func() bool { return false })

	assert.Equal(t, context.Canceled, err)
}
