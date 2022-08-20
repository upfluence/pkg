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
		sf  Singleflight[int32]
		ctr int32
		wg  sync.WaitGroup

		ctx   = context.Background()
		donec = make(chan struct{})
	)

	fn := func(context.Context) (int32, error) {
		res := atomic.AddInt32(&ctr, 1)
		<-donec
		return res, nil
	}

	wg.Add(2)

	go func() {
		ok, res, err := sf.Do(ctx, fn)

		assert.Equal(t, res, int32(1))
		assert.True(t, ok)
		assert.Nil(t, err)

		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		ok, res, err := sf.Do(ctx, fn)

		assert.Equal(t, res, int32(1))
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
		sf  Singleflight[int32]
		ctr int32
		wg  sync.WaitGroup

		ctx   = context.Background()
		donec = make(chan struct{})
	)

	fn := func(context.Context) (int32, error) {
		res := atomic.AddInt32(&ctr, 1)
		<-donec
		return res, nil
	}

	wg.Add(2)

	go func() {
		cctx, cancel := context.WithCancel(ctx)
		cancel()

		ok, _, err := sf.Do(cctx, fn)

		assert.True(t, ok)
		assert.Equal(t, context.Canceled, err)

		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		ok, res, err := sf.Do(ctx, fn)

		assert.Equal(t, res, int32(1))
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
		sf  Singleflight[int32]
		ctr int32
		wg  sync.WaitGroup

		ctx   = context.Background()
		donec = make(chan struct{})
	)

	fn := func(ctx context.Context) (int32, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-donec:
		}

		res := atomic.AddInt32(&ctr, 1)
		return res, nil
	}

	wg.Add(1)

	go func() {
		ok, _, err := sf.Do(ctx, fn)

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
