package syncutil

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSingleflightDedup(t *testing.T) {
	var (
		sf  Singleflight
		ctr int32
		wg  sync.WaitGroup

		ctx   = context.Background()
		donec = make(chan struct{})
	)

	fn := func(context.Context) error {
		atomic.AddInt32(&ctr, 1)
		<-donec
		return nil
	}

	wg.Add(2)

	go func() {
		ok, err := sf.Do(ctx, fn)

		assert.True(t, ok)
		assert.Nil(t, err)

		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		ok, err := sf.Do(ctx, fn)

		assert.False(t, ok)
		assert.Nil(t, err)

		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)
	close(donec)

	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&ctr))

	err := sf.Close()
	assert.Nil(t, err)
}

func TestSingleflightDedupEarlyCancel(t *testing.T) {
	var (
		sf  Singleflight
		ctr int32
		wg  sync.WaitGroup

		ctx   = context.Background()
		donec = make(chan struct{})
	)

	fn := func(context.Context) error {
		atomic.AddInt32(&ctr, 1)
		<-donec
		return nil
	}

	wg.Add(2)

	go func() {
		cctx, cancel := context.WithCancel(ctx)
		cancel()

		ok, err := sf.Do(cctx, fn)

		assert.True(t, ok)
		assert.Equal(t, context.Canceled, err)

		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		ok, err := sf.Do(ctx, fn)

		assert.False(t, ok)
		assert.Nil(t, err)

		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)
	close(donec)

	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&ctr))

	sf.Close()
}

func TestSingleflightStopOnClose(t *testing.T) {
	var (
		sf  Singleflight
		ctr int32
		wg  sync.WaitGroup

		ctx   = context.Background()
		donec = make(chan struct{})
	)

	fn := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-donec:
		}

		atomic.AddInt32(&ctr, 1)
		return nil
	}

	wg.Add(1)

	go func() {
		ok, err := sf.Do(ctx, fn)

		assert.True(t, ok)
		assert.Equal(t, context.Canceled, err)

		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	sf.Close()
	close(donec)

	wg.Wait()

	assert.Equal(t, int32(0), atomic.LoadInt32(&ctr))
}
