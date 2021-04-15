package channelpool

import (
	"context"
	"sync"

	"github.com/streadway/amqp"
	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/amqp/connectionpicker"
	"github.com/upfluence/pkg/closer"
	"github.com/upfluence/pkg/iopool"
	"github.com/upfluence/pkg/log"
)

type Pool interface {
	closer.Shutdowner

	Open(context.Context) error
	IsOpen() bool

	Get(context.Context) (*amqp.Channel, error)
	Put(*amqp.Channel) error
	Discard(*amqp.Channel) error
}

type pool struct {
	*closer.Monitor
	connectionpicker.Picker

	pool *iopool.Pool
	st   sync.Map
}

type Option func(*options)

func WithPoolOptions(opts ...iopool.Option) Option {
	return func(o *options) { o.poopts = append(o.poopts, opts...) }
}

func WithPickerOptions(opts ...connectionpicker.Option) Option {
	return func(o *options) { o.piopts = append(o.piopts, opts...) }
}

func WithPicker(p connectionpicker.Picker) Option {
	return func(o *options) { o.p = p }
}

type options struct {
	poopts []iopool.Option
	piopts []connectionpicker.Option

	p connectionpicker.Picker
}

func NewPool(opts ...Option) Pool {
	var o options

	for _, opt := range opts {
		opt(&o)
	}

	picker := o.p

	if picker == nil {
		picker = connectionpicker.NewPicker(o.piopts...)
	}

	p := pool{Monitor: closer.NewMonitor(), Picker: picker}

	p.pool = iopool.NewPool(p.factory, o.poopts...)

	return &p
}

func (p *pool) IsOpen() bool {
	return p.Picker.IsOpen() && p.Monitor.IsOpen() && p.pool.IsOpen()
}

func (p *pool) Shutdown(ctx context.Context) error {
	return errors.Combine(
		p.pool.Shutdown(ctx),
		p.Monitor.Shutdown(ctx),
		p.Picker.Shutdown(ctx),
	)
}

func (p *pool) Close() error {
	return errors.Combine(
		p.pool.Close(),
		p.Monitor.Close(),
		p.Picker.Close(),
	)
}

type poolEntity struct {
	*amqp.Channel
	err    error
	closed bool
}

func (p *poolEntity) Open(context.Context) error {
	return nil
}

func (p *poolEntity) IsOpen() bool {
	return !p.closed
}

func (p *poolEntity) supervise(ctx context.Context) {
	var ch = make(chan *amqp.Error)

	p.NotifyClose(ch)
	select {
	case err, ok := <-ch:
		if ok {
			if err != nil {
				log.WithError(err).Warning("AMQPChannelPool channel closed")
			}

			p.err = err
		}
		p.closed = true
	case <-ctx.Done():
		p.err = p.Close()
		p.closed = true
	}
}

func (p *pool) factory(ctx context.Context) (iopool.Entity, error) {
	var conn, err = p.Picker.Pick(ctx)

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()

	if err != nil {
		return nil, err
	}

	entity := &poolEntity{Channel: ch}
	p.Run(func(ctx context.Context) { entity.supervise(ctx) })

	return entity, nil
}

func (p *pool) Get(ctx context.Context) (*amqp.Channel, error) {
	var e, err = p.pool.Get(ctx)

	if err != nil {
		return nil, err
	}

	pe := e.(*poolEntity)
	ch := pe.Channel
	p.st.Store(ch, pe)

	return ch, nil
}

func (p *pool) Put(ch *amqp.Channel) error {
	if e, ok := p.st.Load(ch); ok {
		p.st.Delete(ch)
		return p.pool.Put(e.(*poolEntity))
	}

	return nil
}

func (p *pool) Discard(ch *amqp.Channel) error {
	if e, ok := p.st.Load(ch); ok {
		p.st.Delete(ch)
		return p.pool.Discard(e.(*poolEntity))
	}

	return nil
}
