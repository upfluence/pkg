package syncutil

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetryableOnceDo(t *testing.T) {
	var (
		errTransient = errors.New("transient")
		errFatal     = errors.New("fatal")
	)

	for _, tc := range []struct {
		name        string
		opts        []RetryableOnceOption
		haveReturns []error
		wantReturns []error
		wantCalls   int
	}{
		{
			name:        "success caches result",
			haveReturns: []error{nil},
			wantReturns: []error{nil, nil},
			wantCalls:   1,
		},
		{
			name:        "transient error allows retry",
			haveReturns: []error{errTransient, nil},
			wantReturns: []error{errTransient, nil, nil},
			wantCalls:   2,
		},
		{
			name: "fatal error is cached",
			opts: []RetryableOnceOption{
				WithIsFatal(func(err error) bool { return errors.Is(err, errFatal) }),
			},
			haveReturns: []error{errFatal},
			wantReturns: []error{errFatal, errFatal},
			wantCalls:   1,
		},
		{
			name: "non-fatal error retries even with isFatal set",
			opts: []RetryableOnceOption{
				WithIsFatal(func(err error) bool { return errors.Is(err, errFatal) }),
			},
			haveReturns: []error{errTransient, nil},
			wantReturns: []error{errTransient, nil},
			wantCalls:   2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				calls int
				fnIdx int
				ro    = NewRetryableOnce(tc.opts...)
			)

			for i, wantErr := range tc.wantReturns {
				gotErr := ro.Do(func() error {
					ret := tc.haveReturns[fnIdx]
					fnIdx++
					calls++
					return ret
				})

				assert.Equal(t, wantErr, gotErr, "Do call %d", i)
			}

			assert.Equal(t, tc.wantCalls, calls)
		})
	}
}

func TestLazyOnceDo(t *testing.T) {
	var (
		errTransient = errors.New("transient")
		errFatal     = errors.New("fatal")
	)

	for _, tc := range []struct {
		name        string
		opts        []RetryableOnceOption
		haveReturns []struct {
			v   string
			err error
		}
		wantValues []string
		wantErrors []error
		wantCalls  int
	}{
		{
			name: "success caches value",
			haveReturns: []struct {
				v   string
				err error
			}{
				{v: "hello"},
			},
			wantValues: []string{"hello", "hello"},
			wantErrors: []error{nil, nil},
			wantCalls:  1,
		},
		{
			name: "transient error allows retry, zero value returned on error",
			haveReturns: []struct {
				v   string
				err error
			}{
				{err: errTransient},
				{v: "ok"},
			},
			wantValues: []string{"", "ok", "ok"},
			wantErrors: []error{errTransient, nil, nil},
			wantCalls:  2,
		},
		{
			name: "fatal error cached, zero value returned",
			opts: []RetryableOnceOption{
				WithIsFatal(func(err error) bool { return errors.Is(err, errFatal) }),
			},
			haveReturns: []struct {
				v   string
				err error
			}{
				{err: errFatal},
			},
			wantValues: []string{"", ""},
			wantErrors: []error{errFatal, errFatal},
			wantCalls:  1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				calls int
				fnIdx int
				lo    = NewLazyOnce[string](tc.opts...)
			)

			for i, wantErr := range tc.wantErrors {
				got, gotErr := lo.Do(func() (string, error) {
					ret := tc.haveReturns[fnIdx]
					fnIdx++
					calls++
					return ret.v, ret.err
				})

				assert.Equal(t, wantErr, gotErr, "Do call %d", i)
				assert.Equal(t, tc.wantValues[i], got, "Do call %d", i)
			}

			assert.Equal(t, tc.wantCalls, calls)
		})
	}
}

func TestRetryableOnceConcurrent(t *testing.T) {
	var (
		ro    RetryableOnce
		calls int32
		wg    sync.WaitGroup
		ready = make(chan struct{})
	)

	const goroutines = 50

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			<-ready

			err := ro.Do(func() error {
				atomic.AddInt32(&calls, 1)
				return nil
			})

			assert.Nil(t, err)
		}()
	}

	close(ready)
	wg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}
