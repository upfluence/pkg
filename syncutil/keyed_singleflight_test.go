package syncutil

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKeyedSingleflight(t *testing.T) {
	var (
		ksf KeyedSingleflight[string, int32]
		wg  sync.WaitGroup
		ctr int32

		ctx   = context.Background()
		donec = make(chan struct{})
	)

	wg.Add(3)

	go func() {
		ok, res, err := DoOne(ctx, &ksf, "foo", func(context.Context) (int32, error) {
			res := atomic.AddInt32(&ctr, 1)
			<-donec
			return res, nil
		})

		assert.Equal(t, res, int32(1))
		assert.True(t, ok)
		assert.NoError(t, err)
		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		ok, res, err := DoOne(ctx, &ksf, "bar", func(context.Context) (int32, error) {
			res := atomic.AddInt32(&ctr, 1)
			<-donec
			return res, nil
		})

		assert.Equal(t, res, int32(2))
		assert.True(t, ok)
		assert.NoError(t, err)
		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		ok, res, err := DoOne(ctx, &ksf, "foo", func(context.Context) (int32, error) {
			res := atomic.AddInt32(&ctr, 1)
			<-donec
			return res, nil
		})

		assert.Equal(t, res, int32(1))
		assert.False(t, ok)
		assert.NoError(t, err)
		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)
	close(donec)

	wg.Wait()

	assert.Equal(t, 0, len(ksf.sfs))
}
