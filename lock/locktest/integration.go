package locktest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/lock"
)

type TestCase interface {
	Name() string
	Assert(testing.TB, lock.LockManager)
}

type testCase struct {
	name string
	fn   func(testing.TB, lock.LockManager)
}

func (tc testCase) Name() string                             { return tc.name }
func (tc testCase) Assert(t testing.TB, lm lock.LockManager) { tc.fn(t, lm) }

func assertNoWait(t testing.TB, lm lock.LockManager) {
	var (
		lo  = lm.Lock("foo")
		ctx = context.Background()
	)

	le, err := lo.Acquire(
		ctx,
		lock.AcquireOptions{Deadline: time.Now().Add(10 * time.Second)},
	)

	assert.Nil(t, err)

	defer le.Release(ctx)

	_, err = lo.Acquire(
		ctx,
		lock.AcquireOptions{Deadline: time.Now().Add(time.Second), NoWait: true},
	)

	assert.Equal(t, lock.ErrAlreadyAcquired, err)
}

func assertSync(t testing.TB, lm lock.LockManager) {
	var (
		l1       = lm.Lock("foo")
		l2       = lm.Lock("foo")
		ctx      = context.Background()
		acquired = make(chan struct{})

		wg sync.WaitGroup
	)

	le, err := l1.Acquire(
		ctx,
		lock.AcquireOptions{Deadline: time.Now().Add(time.Second)},
	)

	assert.Nil(t, err)

	wg.Add(1)

	go func() {
		le, err := l2.Acquire(
			ctx,
			lock.AcquireOptions{Deadline: time.Now().Add(2 * time.Second)},
		)
		assert.Nil(t, err)
		close(acquired)
		le.Release(ctx)
		wg.Done()
	}()

	select {
	case <-le.Done():
	case <-acquired:
		t.Error("first lock should expire first")
	}

	<-acquired
	wg.Wait()
}

func assertKeepAlive(t testing.TB, lm lock.LockManager) {
	var (
		lo  = lm.Lock("foo")
		ctx = context.Background()
	)

	le, err := lo.Acquire(
		ctx,
		lock.AcquireOptions{Deadline: time.Now().Add(time.Second)},
	)

	assert.Nil(t, err)

	d := time.Now().Add(1 * time.Second)

	assert.Nil(t, le.KeepAlive(ctx, d))
	assert.Equal(t, d, le.Deadline())

	_, err = lo.Acquire(
		ctx,
		lock.AcquireOptions{Deadline: time.Now().Add(time.Second), NoWait: true},
	)

	assert.Equal(t, lock.ErrAlreadyAcquired, err)

	assert.Nil(t, le.Release(ctx))

	_, err = lo.Acquire(
		ctx,
		lock.AcquireOptions{Deadline: time.Now().Add(time.Second), NoWait: true},
	)

	assert.Nil(t, err)
}

func IntegrationTest(t *testing.T, lfn func(testing.TB) lock.LockManager) {
	for _, tt := range []TestCase{
		testCase{name: "no wait", fn: assertNoWait},
		testCase{name: "sync", fn: assertSync},
		testCase{name: "keep alive", fn: assertKeepAlive},
	} {
		t.Run(tt.Name(), func(t *testing.T) {
			tt.Assert(t, lfn(t))
		})
	}
}
