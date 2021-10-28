package time

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/upfluence/pkg/cache/policy"
)

type element struct {
	key string
	t   int64
}

type Policy struct {
	sync.Mutex

	ttl int64

	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	ks map[string]*list.Element
	l  *list.List

	fn func(string)

	closeOnce sync.Once
	ch        chan string
}

func NewIdlePolicy(ttl time.Duration) *Policy {
	return newPolicy(ttl, func(p *Policy) func(string) { return p.move })
}

func NewLifetimePolicy(ttl time.Duration) *Policy {
	return newPolicy(ttl, func(p *Policy) func(string) { return func(string) {} })
}

func newPolicy(ttl time.Duration, fn func(*Policy) func(string)) *Policy {
	p := Policy{
		ttl: int64(ttl),
		ks:  make(map[string]*list.Element),
		l:   list.New(),
		ch:  make(chan string),
	}

	p.fn = fn(&p)
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.wg.Add(1)

	go p.pump()

	return &p
}

func (p *Policy) now() int64 { return time.Now().UnixNano() }

func (p *Policy) C() <-chan string {
	return p.ch
}

func (p *Policy) insert(k string) {
	if p.l.Len() == 0 {
		p.ks[k] = p.l.PushFront(element{key: k, t: p.now()})
	}

	_, ok := p.ks[k]

	if ok {
		return
	}

	p.ks[k] = p.l.InsertAfter(element{key: k, t: p.now()}, p.l.Back())
}

func (p *Policy) move(k string) {
	e, ok := p.ks[k]

	if !ok {
		return
	}

	ee := e.Value.(element)
	ee.t = p.now()

	p.l.MoveToBack(e)
}

func (p *Policy) evict(k string) {
	e, ok := p.ks[k]

	if !ok {
		return
	}

	p.l.Remove(e)
	delete(p.ks, k)
}

func (p *Policy) Op(k string, op policy.OpType) error {
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

func (p *Policy) pump() {
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

func (p *Policy) cleanup(now int64) {
	p.Lock()
	defer p.Unlock()

	e := p.l.Front()

	if e == nil {
		return
	}

	ee := e.Value.(element)

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
		ee = e.Value.(element)
	}
}

func (p *Policy) Close() error {
	p.cancel()
	p.wg.Wait()

	p.closeOnce.Do(func() { close(p.ch) })
	return nil
}
