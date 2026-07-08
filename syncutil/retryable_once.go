package syncutil

import "sync"

type RetryableOnceOption func(*RetryableOnce)

// WithIsFatal overrides the predicate that determines whether an error returned
// by Do's fn should be cached permanently. When isFatal returns true the error
// is stored and all future Do calls return it immediately without re-executing
// fn. When it returns false (the default for all errors) fn will be retried on
// the next Do call.
func WithIsFatal(fn func(error) bool) RetryableOnceOption {
	return func(ro *RetryableOnce) {
		ro.isFatal = fn
	}
}

func NewRetryableOnce(opts ...RetryableOnceOption) *RetryableOnce {
	ro := &RetryableOnce{}

	for _, o := range opts {
		o(ro)
	}

	return ro
}

// RetryableOnce is like sync.Once but allows fn to be retried when it returns
// a non-fatal error. Once fn returns nil or a fatal error (as determined by the
// isFatal predicate), the result is cached and all subsequent Do calls return
// it immediately.
type RetryableOnce struct {
	mu   sync.Mutex
	done bool
	err  error

	isFatal func(error) bool
}

func (ro *RetryableOnce) Do(fn func() error) error {
	ro.mu.Lock()
	defer ro.mu.Unlock()

	if ro.done {
		return ro.err
	}

	err := fn()

	if err == nil || (ro.isFatal != nil && ro.isFatal(err)) {
		ro.done = true
		ro.err = err
	}

	return err
}

// LazyOnce is the value-returning variant of RetryableOnce. It lazily
// computes a value on first successful Do call and caches it for all
// subsequent calls. Like RetryableOnce, non-fatal errors are retried.
type LazyOnce[T any] struct {
	once RetryableOnce
	v    T
}

func NewLazyOnce[T any](opts ...RetryableOnceOption) *LazyOnce[T] {
	lo := &LazyOnce[T]{}

	for _, o := range opts {
		o(&lo.once)
	}

	return lo
}

func (lo *LazyOnce[T]) Do(fn func() (T, error)) (T, error) {
	err := lo.once.Do(func() error {
		v, err := fn()

		if err == nil {
			lo.v = v
		}

		return err
	})

	return lo.v, err
}
