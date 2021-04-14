package iopool

import (
	"context"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/cache/policy"
	"github.com/upfluence/pkg/closer"
)

var (
	ErrNotCheckout = errors.New("iopool: the entity does not belong to the pool")
	ErrClosed      = errors.New("iopool: the pool is closed")
)

type Factory func(context.Context) (Entity, error)

type Entity interface {
	io.Closer

	Open(context.Context) error
	IsOpen() bool
}

type entityWrapper struct {
	e Entity
	n string

	closed bool
}

type Pool struct {
	*closer.Monitor

	size     int
	idleSize int

	closeOnce sync.Once

	factory Factory
	createc chan struct{}
	poolc   chan *entityWrapper

	mu         sync.Mutex
	checkedout map[Entity]*entityWrapper

	checkout map[string]*entityWrapper
	checkin  map[string]*entityWrapper

	ep      policy.EvictionPolicy
	cnt     uint64
	metrics metrics
}

func NewPool(f Factory, opts ...Option) *Pool {
	o := newOptions(opts...)
	p := Pool{
		size:       o.size,
		idleSize:   o.idleSize,
		Monitor:    closer.NewMonitor(),
		factory:    f,
		createc:    make(chan struct{}, o.size),
		poolc:      make(chan *entityWrapper, o.idleSize),
		checkedout: make(map[Entity]*entityWrapper),
		checkout:   make(map[string]*entityWrapper),
		checkin:    make(map[string]*entityWrapper),
		ep:         o.evictionPolicy(),
	}

	p.metrics = newMetrics(o.scope())
	p.metrics.size.Update(int64(o.size))

	for i := 0; i < o.size; i++ {
		p.createc <- struct{}{}
	}

	p.Run(p.closer)

	return &p
}

func (p *Pool) closer(ctx context.Context) {
	for {
		ch := p.ep.C()

		select {
		case <-p.Context().Done():
			return
		case ewn, ok := <-ch:
			if !ok {
				continue
			}

			p.mu.Lock()
			ew, ok := p.checkout[ewn]

			if ok {
				ew.closed = true
				p.mu.Unlock()
				continue
			}

			ew, ok = p.checkin[ewn]

			if ok {
				ew.closed = true
				delete(p.checkin, ew.n)
				p.mu.Unlock()
				ew.e.Close()
			}
		}
	}
}

func (p *Pool) markOut(ew *entityWrapper) {
	p.mu.Lock()
	p.checkedout[ew.e] = ew
	p.checkout[ew.n] = ew
	p.metrics.checkout.Update(int64(len(p.checkout)))
	p.mu.Unlock()
}

func (p *Pool) getNoWait() (Entity, error) {
	for {
		select {
		case ew, ok := <-p.poolc:
			if !ok {
				return nil, ErrClosed
			}

			if p.checkoutWrapper(ew) {
				return ew.e, nil
			}
		default:
			return nil, nil
		}
	}
}

func (p *Pool) Get(ctx context.Context) (Entity, error) {
	p.metrics.get.Inc()

	if !p.IsOpen() {
		return nil, ErrClosed
	}

	e, err := p.getNoWait()

	if err != nil || e != nil {
		return e, err
	}

	t0 := time.Now()
	checkout := func() {
		p.metrics.getWait.Inc()
		p.metrics.getWaitDuration.Add(time.Since(t0).Milliseconds())
	}

	for {
		select {
		case <-p.Context().Done():
			return nil, ErrClosed
		case <-ctx.Done():
			return nil, ctx.Err()
		case ew, ok := <-p.poolc:
			if !ok {
				return nil, ErrClosed
			}

			if p.checkoutWrapper(ew) {
				checkout()
				return ew.e, nil
			}
		case _, ok := <-p.createc:
			if !ok {
				return nil, ErrClosed
			}

			ew, err := p.dial(ctx)

			if err != nil {
				select {
				case p.createc <- struct{}{}:
				default:
				}

				return nil, err
			}

			p.markOut(ew)
			checkout()

			return ew.e, nil
		}
	}
}

func (p *Pool) checkoutWrapper(ew *entityWrapper) bool {
	if ew.closed || !ew.e.IsOpen() {
		ew.e.Close()
		p.mu.Lock()
		delete(p.checkedout, ew.e)
		delete(p.checkout, ew.n)
		p.mu.Unlock()

		select {
		case p.createc <- struct{}{}:
		default:
		}

		return false
	}

	p.mu.Lock()
	delete(p.checkin, ew.n)
	p.mu.Unlock()
	p.markOut(ew)
	p.ep.Op(ew.n, policy.Evict)
	p.metrics.idle.Update(int64(len(p.poolc)))
	return true
}

func (p *Pool) dial(ctx context.Context) (*entityWrapper, error) {
	e, err := p.factory(ctx)

	if err != nil {
		return nil, err
	}

	if err := e.Open(ctx); err != nil {
		return nil, err
	}

	return &entityWrapper{
		e: e,
		n: strconv.Itoa(int(atomic.AddUint64(&p.cnt, 1))),
	}, nil
}

func (p *Pool) requeue(ew *entityWrapper) error {
	select {
	case p.poolc <- ew:
		p.metrics.idle.Update(int64(len(p.poolc)))
	default:
		select {
		case p.createc <- struct{}{}:
		default:
		}

		return ew.e.Close()
	}

	return nil
}

func (p *Pool) Put(e Entity) error {
	if !e.IsOpen() {
		return p.Discard(e)
	}

	p.metrics.put.Inc()

	p.mu.Lock()
	ew, ok := p.checkedout[e]

	if !ok {
		p.mu.Unlock()
		return ErrNotCheckout
	}

	if ew.closed {
		p.mu.Unlock()
		return p.Discard(e)
	}

	p.checkin[ew.n] = ew
	p.ep.Op(ew.n, policy.Set)
	delete(p.checkedout, ew.e)
	delete(p.checkout, ew.n)
	p.metrics.checkout.Update(int64(len(p.checkout)))
	p.mu.Unlock()

	return p.requeue(ew)
}

func (p *Pool) Discard(e Entity) error {
	p.metrics.discard.Inc()
	p.mu.Lock()
	ew, ok := p.checkedout[e]

	if !ok {
		p.mu.Unlock()
		return ErrNotCheckout
	}

	delete(p.checkedout, ew.e)
	delete(p.checkout, ew.n)
	p.ep.Op(ew.n, policy.Evict)
	p.metrics.checkout.Update(int64(len(p.checkout)))
	p.mu.Unlock()

	err := e.Close()

	select {
	case p.createc <- struct{}{}:
	default:
	}

	return err
}

func (p *Pool) drainPoolChannel() []error {
	var errs []error

	for {
		select {
		case ew := <-p.poolc:
			p.metrics.idle.Update(int64(len(p.poolc)))
			p.ep.Op(ew.n, policy.Evict)

			if err := ew.e.Close(); err != nil {
				errs = append(errs, err)
			}
		default:
			return errs
		}
	}
}

func (p *Pool) Shutdown(ctx context.Context) error {
	if err := p.Monitor.Shutdown(ctx); err != nil {
		return err
	}

	errs := p.drainPoolChannel()

	for len(p.checkedout) > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.createc:
		case ew := <-p.poolc:
			p.metrics.idle.Update(int64(len(p.poolc)))

			if err := ew.e.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.WrapErrors(errs)
}

func (p *Pool) Close() error {
	var err error

	p.closeOnce.Do(func() {
		p.Monitor.Close()
		p.ep.Close()

		errs := p.drainPoolChannel()

		p.mu.Lock()
		s := len(p.checkedout)
		p.mu.Unlock()

		for s > 0 {
			select {
			case <-p.createc:
			case ew := <-p.poolc:
				p.metrics.idle.Update(int64(len(p.poolc)))

				if err := ew.e.Close(); err != nil {
					errs = append(errs, err)
				}
			}

			p.mu.Lock()
			s = len(p.checkedout)
			p.mu.Unlock()
		}

		close(p.createc)
		close(p.poolc)

		err = errors.WrapErrors(errs)
	})

	return err
}
