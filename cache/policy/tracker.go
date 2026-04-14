package policy

import (
	"io"
	"sync"
)

// Tracker is the pure eviction logic without any channel machinery. It is
// intended to be wrapped by [Wrap] which provides the channel, the
// closed-check, and the safe send-outside-lock pattern.
//
// Close must stop any background activity and release resources, but must NOT
// close the eviction channel — Wrap does that after all in-flight Ops
// complete.
type Tracker[K comparable] interface {
	Op(K, OpType) error
	io.Closer
}

// Wrap builds a full [EvictionPolicy] from a Tracker factory.
//
// build receives an evict callback and must return a Tracker. The evict
// callback may be called from within Op (synchronous policies) or from a
// background goroutine (asynchronous policies like time). Either way, Wrap
// ensures the channel send completes before the channel is closed.
func Wrap[K comparable](build func(evict func(K)) Tracker[K]) EvictionPolicy[K] {
	ch := make(chan K, 1)

	return &wrapper[K]{
		ch:    ch,
		inner: build(func(k K) { ch <- k }),
	}
}

type wrapper[K comparable] struct {
	inner Tracker[K]

	closeOnce sync.Once
	mu        sync.RWMutex
	closed    bool

	ch chan K
}

func (w *wrapper[K]) C() <-chan K { return w.ch }

func (w *wrapper[K]) Op(k K, op OpType) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		return ErrClosed
	}

	return w.inner.Op(k, op)
}

func (w *wrapper[K]) Close() error {
	var err error

	w.closeOnce.Do(func() {
		w.mu.Lock()
		w.closed = true
		w.mu.Unlock()

		err = w.inner.Close()

		close(w.ch)
	})

	return err
}
