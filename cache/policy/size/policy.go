package size

import (
	"container/list"
	"context"
	"sync"

	"github.com/upfluence/pkg/v2/cache/policy"
)

type Policy[K comparable] struct {
	sync.Mutex

	size int

	l  *list.List
	ks map[K]*list.Element

	ctx    context.Context
	cancel context.CancelFunc

	fn func(K)

	closeOnce sync.Once
	ch        chan K
}

func NewLRUPolicy[K comparable](size int) *Policy[K] {
	p := Policy[K]{
		l:    list.New(),
		size: size,
		ks:   make(map[K]*list.Element),
		ch:   make(chan K, 1),
	}

	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.fn = p.move

	return &p
}

func (p *Policy[K]) C() <-chan K {
	return p.ch
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

func (p *Policy[K]) move(k K) {
	e, ok := p.ks[k]

	if !ok {
		return
	}

	p.l.MoveToBack(e)
}

func (p *Policy[K]) insert(k K) {
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
		v := p.l.Remove(p.l.Front()).(K)
		delete(p.ks, v)

		select {
		case <-p.ctx.Done():
		case p.ch <- v:
		}
	}
}

func (p *Policy[K]) evict(k K) {
	e, ok := p.ks[k]

	if !ok {
		return
	}

	p.l.Remove(e)
	delete(p.ks, k)
}

func (p *Policy[K]) Close() error {
	p.cancel()
	p.closeOnce.Do(func() { close(p.ch) })
	return nil
}
