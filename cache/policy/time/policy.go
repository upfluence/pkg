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

	ks map[K]*list.Element
	l  *list.List

	fn func(K)

	closeOnce sync.Once
	ch        chan K
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
	}

	p.fn = fn(&p)
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.wg.Add(1)

	go p.pump()

	return &p
}

func (p *Policy[K]) now() int64 { return time.Now().UnixNano() }

func (p *Policy[K]) C() <-chan K {
	return p.ch
}

func (p *Policy[K]) insert(k K) {
	if p.l.Len() == 0 {
		p.ks[k] = p.l.PushFront(element[K]{key: k, t: p.now()})
	}

	_, ok := p.ks[k]

	if ok {
		return
	}

	p.ks[k] = p.l.InsertAfter(element[K]{key: k, t: p.now()}, p.l.Back())
}

func (p *Policy[K]) move(k K) {
	e, ok := p.ks[k]

	if !ok {
		return
	}

	ee := e.Value.(element[K])
	ee.t = p.now()

	p.l.MoveToBack(e)
}

func (p *Policy[K]) evict(k K) {
	e, ok := p.ks[k]

	if !ok {
		return
	}

	p.l.Remove(e)
	delete(p.ks, k)
}

func (p *Policy[K]) Op(k K, op policy.OpType) error {
	if p.ctx.Err() != nil {
		return policy.ErrClosed
	}

	p.Lock()

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

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-time.After(time.Duration(p.ttl)):
			now := p.now()

			p.cleanup(now)
		}
	}
}

func (p *Policy[K]) cleanup(now int64) {
	p.Lock()
	defer p.Unlock()

	e := p.l.Front()

	if e == nil {
		return
	}

	ee := e.Value.(element[K])

	for ee.t+p.ttl < now {
		next := e.Next()
		p.l.Remove(e)
		e = next

		k := ee.key
		select {
		case <-p.ctx.Done():
		case p.ch <- k:
		}

		delete(p.ks, k)

		if e == nil {
			return
		}
		ee = e.Value.(element[K])
	}
}

func (p *Policy[K]) Close() error {
	p.cancel()
	p.wg.Wait()

	p.closeOnce.Do(func() { close(p.ch) })
	return nil
}
