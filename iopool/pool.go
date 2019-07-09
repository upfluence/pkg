package iopool

import (
	"context"
	"errors"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/upfluence/pkg/cache/policy"
	ptime "github.com/upfluence/pkg/cache/policy/time"
	"github.com/upfluence/pkg/closer"
	"github.com/upfluence/pkg/multierror"
	"github.com/upfluence/stats"
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

type poolMetrics struct {
	get     stats.Counter
	put     stats.Counter
	discard stats.Counter

	idleClosed stats.Counter

	idle     stats.Gauge
	checkout stats.Gauge
	size     stats.Gauge
}

func newPoolMetrics(s stats.Scope) poolMetrics {
	return poolMetrics{
		get:        s.Counter("get_total"),
		put:        s.Counter("put_total"),
		discard:    s.Counter("discard_total"),
		idleClosed: s.Counter("entity_idle_closed_total"),
		idle:       s.Gauge("idle"),
		checkout:   s.Gauge("checkout"),
		size:       s.Gauge("size"),
	}
}

type entityWrapper struct {
	e Entity
	n string

	closed bool
}

type Pool struct {
	*options
	*closer.Monitor

	factory Factory
	createc chan struct{}
	poolc   chan *entityWrapper

	mu         sync.Mutex
	checkedout map[Entity]*entityWrapper

	checkout map[string]*entityWrapper
	checkin  map[string]*entityWrapper

	ep      policy.EvictionPolicy
	cnt     uint64
	metrics poolMetrics
}

type options struct {
	size     int
	idleSize int

	eps []policy.EvictionPolicy

	scope stats.Scope
}

type Option func(*options)

func WithIdleTimeout(d time.Duration) Option {
	return func(o *options) {
		o.eps = append(o.eps, ptime.NewIdlePolicy(d))
	}
}

func WithScope(s stats.Scope) Option {
	return func(o *options) { o.scope = s }
}

func WithMaxIdle(s int) Option {
	return func(o *options) {
		o.idleSize = s

		if s > o.size {
			o.size = s
		}
	}
}

func WithSize(s int) Option {
	return func(o *options) {
		o.size = s

		if o.idleSize > s {
			o.idleSize = s
		}
	}
}

var defaultOptions = &options{
	size:     10,
	idleSize: 5,
	scope:    stats.RootScope(stats.NewStaticCollector()),
}

func NewPool(f Factory, opts ...Option) *Pool {
	options := *defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	p := Pool{
		Monitor:    closer.NewMonitor(),
		factory:    f,
		options:    &options,
		createc:    make(chan struct{}, options.size),
		poolc:      make(chan *entityWrapper, options.idleSize),
		checkedout: make(map[Entity]*entityWrapper),
		checkout:   make(map[string]*entityWrapper),
		checkin:    make(map[string]*entityWrapper),
		ep:         policy.CombinePolicies(options.eps...),
	}

	p.metrics = newPoolMetrics(options.scope)

	p.metrics.size.Update(int64(options.size))

	for i := 0; i < options.size; i++ {
		p.createc <- struct{}{}
	}

	p.Run(p.closer)

	return &p
}

func (p *Pool) closer(ctx context.Context) {
	for {
		ch := p.ep.C()

		select {
		case <-p.Ctx.Done():
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

	for {
		select {
		case <-p.Ctx.Done():
			return nil, ErrClosed
		case <-ctx.Done():
			return nil, ctx.Err()
		case ew, ok := <-p.poolc:
			if !ok {
				return nil, ErrClosed
			}

			if p.checkoutWrapper(ew) {
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

			return ew.e, nil
		}
	}
}

func (p *Pool) checkoutWrapper(ew *entityWrapper) bool {
	if ew.closed {
		p.mu.Lock()
		delete(p.checkedout, ew.e)
		delete(p.checkout, ew.n)
		p.mu.Unlock()
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

	return multierror.Wrap(errs)
}

func (p *Pool) Close() error {
	p.Monitor.Close()
	p.ep.Close()

	errs := p.drainPoolChannel()

	for len(p.checkedout) > 0 {
		select {
		case <-p.createc:
		case ew := <-p.poolc:
			p.metrics.idle.Update(int64(len(p.poolc)))

			if err := ew.e.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	close(p.createc)
	close(p.poolc)

	return multierror.Wrap(errs)
}
