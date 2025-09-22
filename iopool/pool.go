package iopool

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/cache/policy"
	"github.com/upfluence/pkg/v2/closer"
)

var (
	ErrNotCheckout = errors.New("iopool: the entity does not belong to the pool")
	ErrClosed      = errors.New("iopool: the pool is closed")

	errEmpty = errors.New("pool empty")
)

type Factory[E Entity] func(context.Context) (E, error)

type Entity interface {
	comparable
	io.Closer

	Open(context.Context) error
	IsOpen() bool
}

type entityWrapper[E Entity] struct {
	e E
	n uint64

	closed bool
}

type Pool[E Entity] struct {
	*closer.Monitor

	size     int
	idleSize int

	closeOnce sync.Once

	factory Factory[E]
	createc chan struct{}
	poolc   chan *entityWrapper[E]

	mu         sync.Mutex
	checkedout map[E]*entityWrapper[E]

	checkout map[uint64]*entityWrapper[E]
	checkin  map[uint64]*entityWrapper[E]

	ep      policy.EvictionPolicy[uint64]
	cnt     uint64
	metrics metrics
}

func NewPool[E Entity](f Factory[E], opts ...Option) *Pool[E] {
	o := newOptions(opts...)
	p := Pool[E]{
		size:       o.size,
		idleSize:   o.idleSize,
		Monitor:    closer.NewMonitor(),
		factory:    f,
		createc:    make(chan struct{}, o.size),
		poolc:      make(chan *entityWrapper[E], o.idleSize),
		checkedout: make(map[E]*entityWrapper[E]),
		checkout:   make(map[uint64]*entityWrapper[E]),
		checkin:    make(map[uint64]*entityWrapper[E]),
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

func (p *Pool[E]) closer(ctx context.Context) {
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

				var zero E
				ew.e = zero
			}
		}
	}
}

func (p *Pool[E]) markOut(ew *entityWrapper[E]) {
	p.mu.Lock()
	p.checkedout[ew.e] = ew
	p.checkout[ew.n] = ew
	p.metrics.checkout.Update(int64(len(p.checkout)))
	p.mu.Unlock()
}

func (p *Pool[E]) getNoWait() (E, error) {
	var zero E

	for {
		select {
		case ew, ok := <-p.poolc:
			if !ok {
				return zero, ErrClosed
			}

			if p.checkoutWrapper(ew) {
				return ew.e, nil
			}
		default:
			return zero, errEmpty
		}
	}
}

func (p *Pool[E]) Get(ctx context.Context) (E, error) {
	var zero E

	p.metrics.get.Inc()

	if !p.IsOpen() {
		return zero, ErrClosed
	}

	e, err := p.getNoWait()

	if err != errEmpty {
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
			return zero, ErrClosed
		case <-ctx.Done():
			return zero, ctx.Err()
		case ew, ok := <-p.poolc:
			if !ok {
				return zero, ErrClosed
			}

			if p.checkoutWrapper(ew) {
				checkout()
				return ew.e, nil
			}
		case _, ok := <-p.createc:
			if !ok {
				return zero, ErrClosed
			}

			ew, err := p.dial(ctx)

			if err != nil {
				select {
				case p.createc <- struct{}{}:
				default:
				}

				return zero, err
			}

			p.markOut(ew)
			checkout()

			return ew.e, nil
		}
	}
}

func (p *Pool[E]) checkoutWrapper(ew *entityWrapper[E]) bool {
	var zero E

	if ew.closed || ew.e == zero || !ew.e.IsOpen() {
		if ew.e != zero {
			ew.e.Close()
		}
		p.mu.Lock()

		if ew.e != zero {
			delete(p.checkedout, ew.e)
		}

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

func (p *Pool[E]) dial(ctx context.Context) (*entityWrapper[E], error) {
	e, err := p.factory(ctx)

	if err != nil {
		return nil, err
	}

	if err := e.Open(ctx); err != nil {
		return nil, err
	}

	return &entityWrapper[E]{
		e: e,
		n: atomic.AddUint64(&p.cnt, 1),
	}, nil
}

func (p *Pool[E]) requeue(ew *entityWrapper[E]) error {
	select {
	case p.poolc <- ew:
		p.metrics.idle.Update(int64(len(p.poolc)))
	default:
		select {
		case p.createc <- struct{}{}:
		default:
		}

		p.ep.Op(ew.n, policy.Evict)
		p.mu.Lock()
		delete(p.checkin, ew.n)
		p.mu.Unlock()
		return ew.e.Close()
	}

	return nil
}

func (p *Pool[E]) Put(e E) error {
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
	delete(p.checkedout, ew.e)
	delete(p.checkout, ew.n)
	p.mu.Unlock()

	p.ep.Op(ew.n, policy.Set)
	p.metrics.checkout.Update(int64(len(p.checkout)))

	return p.requeue(ew)
}

func (p *Pool[E]) Discard(e E) error {
	p.metrics.discard.Inc()
	p.mu.Lock()
	ew, ok := p.checkedout[e]

	if !ok {
		p.mu.Unlock()
		return ErrNotCheckout
	}

	delete(p.checkedout, ew.e)
	delete(p.checkout, ew.n)
	p.mu.Unlock()

	p.ep.Op(ew.n, policy.Evict)
	p.metrics.checkout.Update(int64(len(p.checkout)))

	err := e.Close()

	select {
	case p.createc <- struct{}{}:
	default:
	}

	return err
}

func (p *Pool[E]) drainPoolChannel() []error {
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

func (p *Pool[E]) Shutdown(ctx context.Context) error {
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

func (p *Pool[E]) Close() error {
	var (
		err  error
		zero E
	)

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

				if !ew.closed && ew.e != zero {
					if err := ew.e.Close(); err != nil {
						errs = append(errs, err)
					}
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
