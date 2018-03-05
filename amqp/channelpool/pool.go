package channelpool

import (
	"context"

	"github.com/streadway/amqp"

	"github.com/upfluence/pkg/amqp/connectionpicker"
	"github.com/upfluence/pkg/log"
	stdpool "github.com/upfluence/pkg/pool"
)

type PoolFactory interface {
	GetPool(stdpool.PoolFactory, connectionpicker.Picker) Pool
}

type Pool interface {
	Open(context.Context) error
	IsOpen() bool
	Close() error

	Get(context.Context) (*amqp.Channel, error)
	Put(*amqp.Channel) error
	Discard(*amqp.Channel) error
}

func NewPool(f stdpool.PoolFactory, picker connectionpicker.Picker) Pool {
	p := &pool{Picker: picker, st: map[*amqp.Channel]*poolEntity{}}
	p.pool = f.GetPool(p.factory)

	return p
}

type pool struct {
	connectionpicker.Picker

	pool stdpool.Pool
	st   map[*amqp.Channel]*poolEntity
}

type poolEntity struct {
	channel *amqp.Channel
	opened  bool
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

func (p *pool) factory(ctx context.Context) (interface{}, error) {
	var conn, err = p.Picker.Pick(ctx)

	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()

	if err != nil {
		return nil, err
	}

	entity := &poolEntity{channel: ch, opened: true}
	go entity.supervise()

	return entity, nil
}

func (p *pool) Get(ctx context.Context) (*amqp.Channel, error) {
	var e, err = p.pool.Get(ctx)

	if err != nil {
		return nil, err
	}

	for !e.(*poolEntity).opened {
		if err2 := p.pool.Discard(e); err2 != nil {
			log.Errorf("amqputil: %v", err2)
		}

		e, err = p.pool.Get(ctx)

		if err != nil {
			return nil, err
		}
	}

	p.st[e.(*poolEntity).channel] = e.(*poolEntity)
	return e.(*poolEntity).channel, nil
}

func (p *pool) Put(ch *amqp.Channel) error {
	if e, ok := p.st[ch]; ok {
		delete(p.st, ch)

		if e.opened {
			return p.pool.Put(e)
		}

		return p.pool.Discard(e)
	}

	return nil
}

func (p *pool) Discard(ch *amqp.Channel) error {
	if e, ok := p.st[ch]; ok {
		delete(p.st, ch)

		if e.opened {
			if err := ch.Close(); err != nil {
				log.Errorf("amqputil: %v", err)
			}
		}

		return p.pool.Discard(e)
	}

	return nil
}
