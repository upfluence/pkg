package bounded

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

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
	pool     chan interface{}
	poolSize int32
	limit    int

	checkoutL *sync.Mutex
	checkout  map[interface{}]bool

	createChannel chan interface{}

	factory pool.Factory
}

func NewPoolFactory(size int) *PoolFactory {
	return &PoolFactory{size: size}
}

func (f *PoolFactory) GetIntrospectablePool(factory pool.Factory) pool.IntrospectablePool {
	return NewPool(f.size, factory)
}

func (f *PoolFactory) GetPool(factory pool.Factory) pool.Pool {
	return f.GetIntrospectablePool(factory)
}

func NewPool(limit int, factory pool.Factory) *Pool {
	var p = &Pool{
		pool:          make(chan interface{}, limit),
		createChannel: make(chan interface{}, limit),
		limit:         limit,
		checkout:      make(map[interface{}]bool),
		checkoutL:     &sync.Mutex{},
		poolSize:      0,
		factory:       factory,
	}

	for i := 0; i < limit; i++ {
		p.createChannel <- true
	}

	return p
}

func (p *Pool) checkin(e interface{}) {
	p.checkoutL.Lock()
	defer p.checkoutL.Unlock()

	p.checkout[e] = true
}

func (p *Pool) Get(ctx context.Context) (interface{}, error) {
	// Prefer entity from pool rather than a new one
	select {
	case e := <-p.pool:
		p.checkin(e)

		atomic.AddInt32(&p.poolSize, -1)

		return e, nil
	default:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case e := <-p.pool:
		p.checkin(e)

		atomic.AddInt32(&p.poolSize, -1)

		return e, nil
	case <-p.createChannel:
		e, err := p.factory(ctx)

		if err != nil {
			select {
			case p.createChannel <- true:
			default:
			}

			return nil, err
		}

		p.checkin(e)
		return e, nil
	}
}

func (p *Pool) Put(e interface{}) error {
	p.checkoutL.Lock()
	defer p.checkoutL.Unlock()

	if _, ok := p.checkout[e]; !ok {
		return errNotCheckout
	}

	delete(p.checkout, e)

	select {
	case p.pool <- e:
		atomic.AddInt32(&p.poolSize, 1)
	default:
	}

	return nil
}

func (p *Pool) GetStats() (int, int) {
	p.checkoutL.Lock()
	defer p.checkoutL.Unlock()

	return int(p.poolSize), len(p.checkout)
}

func (p *Pool) Discard(e interface{}) error {
	p.checkoutL.Lock()
	defer p.checkoutL.Unlock()

	if _, ok := p.checkout[e]; !ok {
		return errNotCheckout
	}

	delete(p.checkout, e)

	select {
	case p.createChannel <- true:
	default:
	}

	return nil
}
