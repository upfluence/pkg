package amqputil

import (
	"context"
	"errors"
	"time"

	"github.com/streadway/amqp"

	"github.com/upfluence/pkg/backoff"
	"github.com/upfluence/pkg/backoff/static"
	"github.com/upfluence/pkg/log"
	"github.com/upfluence/pkg/pool"
	"github.com/upfluence/pkg/pool/bounded"
)

var (
	ErrCancelled = errors.New("amqputil: Retry cancelled")

	defaultBackoff = static.NewBackoff(5, 200*time.Millisecond)
)

type ChannelPool struct {
	pool    pool.Pool
	backoff backoff.Strategy

	picker ConnectionPicker
	st     map[*amqp.Channel]*poolEntity
}

func NewChannelPoolFromAmqpURL(amqpURL string, size int) (*ChannelPool, error) {
	var picker, err = NewSingleConnectionPicker(amqpURL)

	if err != nil {
		return nil, err
	}

	p := &ChannelPool{
		backoff: defaultBackoff,
		picker:  picker,
		st:      make(map[*amqp.Channel]*poolEntity),
	}

	p.pool = bounded.NewPool(size, p.factory)

	return p, nil
}

type poolEntity struct {
	channel  *amqp.Channel
	callback func()
	opened   bool
}

func (p *poolEntity) supervise() {
	var ch = make(chan *amqp.Error)

	p.channel.NotifyClose(ch)

	err := <-ch

	if err != nil {
		log.Errorf("AMQPChannelPool channel closed: %v", err)
	}

	p.opened = false
}

func (p *ChannelPool) factory(ctx context.Context) (interface{}, error) {
	var conn, callback, err = p.picker.Pick(ctx)

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()

	if err != nil {
		return nil, err
	}

	entity := &poolEntity{channel: ch, callback: callback, opened: true}
	go entity.supervise()

	return entity, nil
}

func (p *ChannelPool) Get(ctx context.Context) (*amqp.Channel, error) {
	var (
		i int

		e, err = p.pool.Get(ctx)
	)

	if err != nil {
		return nil, err
	}

	for !e.(*poolEntity).opened {
		e.(*poolEntity).callback()
		if err := p.pool.Discard(e); err != nil {
			log.Errorf("amqputil: %v", err)
		}

		d, err := p.backoff.Backoff(i)

		if err != nil {
			return nil, err
		}

		if d == backoff.Cancelled {
			return nil, ErrCancelled
		}

		time.Sleep(d)

		e, err = p.pool.Get(ctx)

		if err != nil {
			return nil, err
		}

		i++
	}

	p.st[e.(*poolEntity).channel] = e.(*poolEntity)
	return e.(*poolEntity).channel, nil
}

func (p *ChannelPool) Put(ch *amqp.Channel) error {
	if e, ok := p.st[ch]; ok {
		delete(p.st, ch)

		if e.opened {
			return p.pool.Put(e)
		}

		e.callback()
		return p.pool.Discard(e)
	}

	return nil
}

func (p *ChannelPool) Discard(ch *amqp.Channel) error {
	if e, ok := p.st[ch]; ok {
		delete(p.st, ch)

		if e.opened {
			if err := ch.Close(); err != nil {
				log.Errorf("amqputil: %v", err)
			}
		}

		e.callback()
		return p.pool.Discard(e)
	}

	return nil
}
