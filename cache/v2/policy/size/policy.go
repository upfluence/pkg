package size

import (
	"container/list"
	"context"
	"sync"

	"github.com/upfluence/pkg/cache/v2/policy"
)

type Policy struct {
	sync.Mutex

	size int

	l  *list.List
	ks map[string]*list.Element

	ctx    context.Context
	cancel context.CancelFunc

	fn func(string)

	closeOnce sync.Once
	ch        chan string
}

func NewLRUPolicy(size int) *Policy {
	p := Policy{
		l:    list.New(),
		size: size,
		ks:   make(map[string]*list.Element),
		ch:   make(chan string, 1),
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.fn = p.move

	return &p
}

func (p *Policy) C() <-chan string {
	return p.ch
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

func (p *Policy) move(k string) {
	e, ok := p.ks[k]

	if !ok {
		return
	}

	p.l.MoveToBack(e)
}

func (p *Policy) insert(k string) {
	if _, ok := p.ks[k]; ok {
		return
	}

	var e *list.Element

	if p.l.Len() == 0 {
		e = p.l.PushFront(k)
	} else {
		e = p.l.InsertAfter(k, p.l.Back())
	}

	p.ks[k] = e

	if p.l.Len() > p.size {
		v := p.l.Remove(p.l.Front()).(string)
		delete(p.ks, v)

		select {
		case <-p.ctx.Done():
		case p.ch <- v:
		}
	}
}

func (p *Policy) evict(k string) {
	e, ok := p.ks[k]

	if !ok {
		return
	}

	p.l.Remove(e)
	delete(p.ks, k)
}

func (p *Policy) Close() error {
	p.cancel()
	p.closeOnce.Do(func() { close(p.ch) })
	return nil
}
