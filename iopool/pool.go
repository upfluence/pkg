package iopool

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

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
	p *Pool

	lastIdle time.Time
}

func (ew *entityWrapper) valid() bool {
	t := time.Now()

	if ew.p.idleTimeout > 0 && ew.lastIdle.Before(t.Add(-1*ew.p.idleTimeout)) {
		return false
	}

	return ew.e.IsOpen()
}

type Pool struct {
	*options

	wg sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	closed int32

	ticker *time.Ticker

	factory  Factory
	createCh chan struct{}
	poolCh   chan *entityWrapper

	checkoutMu sync.Mutex
	checkout   map[Entity]*entityWrapper

	metrics poolMetrics
}

type options struct {
	size     int
	idleSize int

	idleTimeout    time.Duration
	closerInterval time.Duration

	scope stats.Scope
}

type Option func(*options)

func WithIdleTimeout(d time.Duration) Option {
	return func(o *options) {
		o.idleTimeout = d

		if d < o.closerInterval {
			o.closerInterval = d
		}
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

var defaultOptions = options{
	size:           10,
	idleSize:       5,
	closerInterval: time.Second,
	scope:          stats.RootScope(stats.NewStaticCollector()),
}

func NewPool(f Factory, opts ...Option) *Pool {
	var (
		options     = defaultOptions
		ctx, cancel = context.WithCancel(context.Background())
	)

	for _, opt := range opts {
		opt(&options)
	}

	p := &Pool{
		ctx:      ctx,
		cancel:   cancel,
		factory:  f,
		options:  &options,
		createCh: make(chan struct{}, options.size),
		poolCh:   make(chan *entityWrapper, options.idleSize),
		checkout: make(map[Entity]*entityWrapper),
	}

	p.metrics = newPoolMetrics(options.scope)

	p.metrics.size.Update(int64(options.size))

	if options.idleTimeout > 0 {
		p.ticker = time.NewTicker(options.closerInterval)
		p.wg.Add(1)
		go p.closer()
	}

	for i := 0; i < options.size; i++ {
		p.createCh <- struct{}{}
	}

	return p
}

func (p *Pool) closer() {
	defer p.wg.Done()
	defer p.ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.ticker.C:
			var ews []*entityWrapper

			for {
				select {
				case ew := <-p.poolCh:
					if ew.valid() {
						ews = append(ews, ew)
						continue
					}

					p.metrics.idleClosed.Inc()
					ew.e.Close()
				default:
					goto reenqueue
				}
			}

		reenqueue:
			for _, ew := range ews {
				p.poolCh <- ew
			}

			p.metrics.idle.Update(int64(len(p.poolCh)))
		}
	}
}

func (p *Pool) markOut(ew *entityWrapper) {
	p.checkoutMu.Lock()
	p.checkout[ew.e] = ew
	p.metrics.checkout.Update(int64(len(p.checkout)))
	p.checkoutMu.Unlock()
}

func (p *Pool) IsOpen() bool {
	return p.closed == 0
}

func (p *Pool) Get(ctx context.Context) (Entity, error) {
	p.metrics.get.Inc()

	if atomic.LoadInt32(&p.closed) == 1 {
		return nil, ErrClosed
	}

	select {
	case e, ok := <-p.poolCh:
		if !ok {
			return nil, ErrClosed
		}

		return e.e, nil
	default:
	}

	select {
	case <-p.ctx.Done():
		return nil, ErrClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	case ew, ok := <-p.poolCh:
		if !ok {
			return nil, ErrClosed
		}

		p.markOut(ew)
		p.metrics.idle.Update(int64(len(p.poolCh)))

		return ew.e, nil
	case _, ok := <-p.createCh:
		if !ok {
			return nil, ErrClosed
		}

		e, err := p.factory(ctx)

		if err != nil {
			select {
			case p.createCh <- struct{}{}:
			default:
			}

			return nil, err
		}

		if err := e.Open(ctx); err != nil {
			select {
			case p.createCh <- struct{}{}:
			default:
			}

			return nil, err
		}

		t := time.Now()
		p.markOut(&entityWrapper{e: e, lastIdle: t, p: p})

		return e, nil
	}
}

func (p *Pool) requeue(ew *entityWrapper) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ew.e.Close()
	}

	select {
	case p.poolCh <- ew:
		p.metrics.idle.Update(int64(len(p.poolCh)))
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

	p.checkoutMu.Lock()
	ew, ok := p.checkout[e]

	if !ok {
		p.checkoutMu.Unlock()
		return ErrNotCheckout
	}

	ew.lastIdle = time.Now()

	delete(p.checkout, ew.e)
	p.metrics.checkout.Update(int64(len(p.checkout)))
	p.checkoutMu.Unlock()

	return p.requeue(ew)
}

func (p *Pool) Discard(e Entity) error {
	p.metrics.discard.Inc()
	p.checkoutMu.Lock()
	ew, ok := p.checkout[e]

	if !ok {
		p.checkoutMu.Unlock()
		return ErrNotCheckout
	}

	delete(p.checkout, ew.e)
	p.metrics.checkout.Update(int64(len(p.checkout)))
	p.checkoutMu.Unlock()

	err := e.Close()

	if atomic.LoadInt32(&p.closed) == 0 {
		select {
		case p.createCh <- struct{}{}:
		default:
		}
	}

	return err
}

func (p *Pool) drainPoolChannel() []error {
	var errs []error

	for {
		select {
		case ew := <-p.poolCh:
			p.metrics.idle.Update(int64(len(p.poolCh)))

			if err := ew.e.Close(); err != nil {
				errs = append(errs, err)
			}
		default:
			return errs
		}
	}
}

func (p *Pool) Close() error {
	p.cancel()
	p.wg.Wait()

	errs := p.drainPoolChannel()

	for len(p.checkout) > 0 {
		select {
		case <-p.createCh:
		case ew := <-p.poolCh:
			p.metrics.idle.Update(int64(len(p.poolCh)))

			if err := ew.e.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	close(p.createCh)
	close(p.poolCh)

	atomic.StoreInt32(&p.closed, 1)

	return multierror.Wrap(errs)
}
