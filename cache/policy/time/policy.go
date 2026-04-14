package time

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/upfluence/pkg/v2/cache/policy"
)

type element[K comparable] struct {
	key K
	t   int64
}

type Policy[K comparable] struct {
	sync.Mutex

	ttl int64

	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	// closed is protected by the embedded Mutex (same lock used by Op/cleanup).
	closed bool

	// ks maps each key to its *list.Element, whose Value is *element[K].
	// Storing a pointer means move() can update the timestamp in place.
	ks map[K]*list.Element
	l  *list.List

	// fn is called on Get: no-op for lifetime policy, move for idle policy.
	fn func(K)

	// now is overridable in tests; defaults to time.Now().UnixNano().
	now func() int64

	// sends tracks goroutines blocked on a channel send. Close() waits for
	// all of them to finish before closing the channel.
	sends sync.WaitGroup

	ch chan K
}

func NewIdlePolicy[K comparable](ttl time.Duration) *Policy[K] {
	return newPolicy[K](ttl, func(p *Policy[K]) func(K) { return p.move })
}

func NewLifetimePolicy[K comparable](ttl time.Duration) *Policy[K] {
	return newPolicy[K](ttl, func(p *Policy[K]) func(K) { return func(K) {} })
}

func newPolicy[K comparable](ttl time.Duration, fn func(*Policy[K]) func(K)) *Policy[K] {
	p := Policy[K]{
		ttl: int64(ttl),
		ks:  make(map[K]*list.Element),
		l:   list.New(),
		ch:  make(chan K),
		now: func() int64 { return time.Now().UnixNano() },
	}

	p.fn = fn(&p)
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.wg.Add(1)

	go p.pump()

	return &p
}

func (p *Policy[K]) C() <-chan K {
	return p.ch
}

// insert adds k to the ordered list. Must be called with p.Mutex held.
func (p *Policy[K]) insert(k K) {
	if _, ok := p.ks[k]; ok {
		return
	}

	e := p.l.PushBack(&element[K]{key: k, t: p.now()})
	p.ks[k] = e
}

// move refreshes the timestamp and moves k to the back (idle-policy Get).
// Must be called with p.Mutex held.
func (p *Policy[K]) move(k K) {
	e, ok := p.ks[k]
	if !ok {
		return
	}

	// e.Value is *element[K] — update in place, no copy.
	e.Value.(*element[K]).t = p.now()
	p.l.MoveToBack(e)
}

// evict removes k from the list and map. Must be called with p.Mutex held.
func (p *Policy[K]) evict(k K) {
	e, ok := p.ks[k]
	if !ok {
		return
	}

	p.l.Remove(e)
	delete(p.ks, k)
}

func (p *Policy[K]) Op(k K, op policy.OpType) error {
	p.Lock()

	if p.closed {
		p.Unlock()
		return policy.ErrClosed
	}

	switch op {
	case policy.Set:
		p.insert(k)
	case policy.Get:
		p.fn(k)
	case policy.Evict:
		p.evict(k)
	}

	p.Unlock()
	return nil
}

func (p *Policy[K]) pump() {
	defer p.wg.Done()

	t := time.NewTimer(time.Duration(p.ttl))
	defer t.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-t.C:
			p.cleanup(p.now())
			t.Reset(time.Duration(p.ttl))
		}
	}
}

// cleanup collects expired keys under the lock, registers them with sends,
// then sends them on p.ch without holding the lock. Close() waits for all
// registered sends to complete before closing the channel.
func (p *Policy[K]) cleanup(now int64) {
	p.Lock()

	var toEvict []K

	for e := p.l.Front(); e != nil; {
		ee := e.Value.(*element[K])
		if ee.t+p.ttl >= now {
			break
		}

		next := e.Next()
		p.l.Remove(e)
		delete(p.ks, ee.key)
		toEvict = append(toEvict, ee.key)
		e = next
	}

	if len(toEvict) > 0 {
		// Register all sends before releasing the lock so Close() cannot
		// observe sends==0 between Unlock and the sends below.
		p.sends.Add(len(toEvict))
	}

	p.Unlock()

	for i, k := range toEvict {
		select {
		case <-p.ctx.Done():
			// Context cancelled: mark the remaining sends (including this
			// one) as done so Close()'s sends.Wait() is not blocked.
			for range toEvict[i:] {
				p.sends.Done()
			}
			return
		case p.ch <- k:
			p.sends.Done()
		}
	}
}

func (p *Policy[K]) Close() error {
	p.Lock()
	p.closed = true
	p.Unlock()

	p.cancel()
	p.wg.Wait()

	// Wait for any sends registered before closed=true / cancel() to finish.
	p.sends.Wait()

	close(p.ch)
	return nil
}
