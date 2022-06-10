package syncutil

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKeyedSingleflight(t *testing.T) {
	var (
		ksf KeyedSingleflight[string]
		wg  sync.WaitGroup

		ctx   = context.Background()
		donec = make(chan struct{})
	)

	wg.Add(3)

	go func() {
		ok, _ := ksf.Do(ctx, "foo", func(context.Context) error {
			<-donec
			return nil
		})

		assert.True(t, ok)
		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		ok, _ := ksf.Do(ctx, "bar", func(context.Context) error {
			<-donec
			return nil
		})

		assert.True(t, ok)
		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)

	go func() {
		ok, _ := ksf.Do(ctx, "foo", func(context.Context) error {
			<-donec
			return nil
		})

		assert.False(t, ok)
		wg.Done()
	}()

	time.Sleep(10 * time.Millisecond)
	close(donec)

	wg.Wait()

	assert.Equal(t, 0, len(ksf.sfs))
}
