package time

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/upfluence/pkg/v2/cache/policy"
	"github.com/upfluence/pkg/v2/timeutil"
)

type element[K comparable] struct {
	key K
	t   time.Time
}

// NewIdlePolicy returns a TTL-based eviction policy that resets the clock on
// every Get (idle / sliding-window expiry).
func NewIdlePolicy[K comparable](ttl time.Duration) policy.EvictionPolicy[K] {
	return newPolicy[K](ttl, func(t *tracker[K]) func(K) { return t.move })
}

// NewLifetimePolicy returns a TTL-based eviction policy that expires keys a
// fixed duration after they were first Set, regardless of access.
func NewLifetimePolicy[K comparable](ttl time.Duration) policy.EvictionPolicy[K] {
	return newPolicy[K](ttl, func(*tracker[K]) func(K) { return func(K) {} })
}

func newPolicy[K comparable](ttl time.Duration, fn func(*tracker[K]) func(K)) policy.EvictionPolicy[K] {
	return policy.Wrap(func(evict func(K)) policy.Tracker[K] {
		return newTracker[K](ttl, fn, evict, timeutil.Background())
	})
}

// newTracker constructs a bare tracker without wrapping it. Used by tests to
// inject a fake clock and call cleanup directly.
func newTracker[K comparable](ttl time.Duration, fn func(*tracker[K]) func(K), evict func(K), clock timeutil.Clock) *tracker[K] {
	t := &tracker[K]{
		ttl:   ttl,
		ks:    make(map[K]*list.Element),
		l:     list.New(),
		clock: clock,
		evict: evict,
	}

	t.fn = fn(t)
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.wg.Add(1)

	go t.pump()

	return t
}

type tracker[K comparable] struct {
	sync.Mutex

	ttl time.Duration

	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	// ks maps each key to its *list.Element, whose Value is *element[K].
	ks map[K]*list.Element
	l  *list.List

	// fn is called on Get: no-op for lifetime, move for idle.
	fn func(K)

	clock timeutil.Clock

	// evict is the callback supplied by Wrap; calling it sends k on the
	// eviction channel in a goroutine-safe way.
	evict func(K)
}

func (t *tracker[K]) Op(k K, op policy.OpType) error {
	t.Lock()
	defer t.Unlock()

	switch op {
	case policy.Set:
		t.insert(k)
	case policy.Get:
		t.fn(k)
	case policy.Evict:
		t.remove(k)
	}

	return nil
}

// Close stops the background pump goroutine and waits for it to exit.
func (t *tracker[K]) Close() error {
	t.cancel()
	t.wg.Wait()
	return nil
}

func (t *tracker[K]) insert(k K) {
	if _, ok := t.ks[k]; ok {
		return
	}

	e := t.l.PushBack(&element[K]{key: k, t: t.clock.Now()})
	t.ks[k] = e
}

func (t *tracker[K]) move(k K) {
	e, ok := t.ks[k]
	if !ok {
		return
	}

	e.Value.(*element[K]).t = t.clock.Now()
	t.l.MoveToBack(e)
}

func (t *tracker[K]) remove(k K) {
	e, ok := t.ks[k]
	if !ok {
		return
	}

	t.l.Remove(e)
	delete(t.ks, k)
}

func (t *tracker[K]) pump() {
	defer t.wg.Done()

	timer := t.clock.Timer(t.ttl)
	defer timer.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-timer.C():
			t.cleanup()
			timer.Reset(t.ttl)
		}
	}
}

// cleanup collects expired keys under the lock, then sends them via the evict
// callback outside the lock.
func (t *tracker[K]) cleanup() {
	now := t.clock.Now()

	var toEvict []K

	t.Lock()

	for e := t.l.Front(); e != nil; {
		ee := e.Value.(*element[K])
		if !ee.t.Add(t.ttl).Before(now) {
			break
		}

		next := e.Next()
		t.l.Remove(e)
		delete(t.ks, ee.key)
		toEvict = append(toEvict, ee.key)
		e = next
	}

	t.Unlock()

	for _, k := range toEvict {
		select {
		case <-t.ctx.Done():
			return
		default:
			t.evict(k)
		}
	}
}
