package bounded

import (
	"context"
	"errors"
	"sync"

	"github.com/upfluence/pkg/pool"
)

var (
	errPoolFull    = errors.New("pool/bounded: pool full")
	errNotCheckout = errors.New("pool/bounded: the entity does not belong to the pool")
)

type PoolFactory struct {
	size int
}

type Pool struct {
	pool  chan interface{}
	cond  *sync.Cond
	limit int

	checkout map[interface{}]bool

	factory pool.Factory
}

func NewPoolFactory(size int) *PoolFactory {
	return &PoolFactory{size: size}
}

func (f *PoolFactory) GetPool(factory pool.Factory) pool.Pool {
	return NewPool(f.size, factory)
}

func NewPool(limit int, factory pool.Factory) *Pool {
	return &Pool{
		pool:     make(chan interface{}, limit),
		cond:     sync.NewCond(&sync.Mutex{}),
		limit:    limit,
		checkout: make(map[interface{}]bool),
		factory:  factory,
	}
}

func (p *Pool) Get(ctx context.Context) (interface{}, error) {
	select {
	case e := <-p.pool:
		p.cond.L.Lock()
		defer p.cond.L.Unlock()
		p.checkout[e] = true
		return e, nil
	default:
	}

	var (
		cancelled bool

		ch = make(chan bool)
	)

	go func() {
		p.cond.L.Lock()
		defer p.cond.L.Unlock()

		for len(p.checkout) >= p.limit {
			p.cond.Wait()

			if cancelled {
				p.cond.Signal()
				return
			}
		}

		ch <- true
	}()

	select {
	case <-ctx.Done():
		cancelled = true
		return nil, ctx.Err()
	case e := <-p.pool:
		cancelled = true
		p.checkout[e] = true
		return e, nil
	case <-ch:
		e, err := p.factory(ctx)

		if err == nil {
			p.cond.L.Lock()
			p.checkout[e] = true
			p.cond.L.Unlock()
		}

		return e, err
	}
}

func (p *Pool) Put(e interface{}) error {
	if _, ok := p.checkout[e]; !ok {
		return errNotCheckout
	}

	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	delete(p.checkout, e)

	select {
	case p.pool <- e:
		return nil
	default:
		return errPoolFull
	}
}

func (p *Pool) Discard(e interface{}) error {
	if _, ok := p.checkout[e]; !ok {
		return errNotCheckout
	}

	p.cond.L.Lock()
	defer p.cond.L.Unlock()

	delete(p.checkout, e)

	p.cond.Signal()

	return nil
}
