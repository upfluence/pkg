package amqputil

import (
	"context"
	"sync"

	"github.com/streadway/amqp"
	"github.com/upfluence/pkg/log"
)

type ConnectionPicker interface {
	Pick(context.Context) (*amqp.Connection, func(), error)
}

type SingleConnectionPicker struct {
	amqpURL string
	closed  bool

	conn *amqp.Connection
	cond *sync.Cond
}

func NewSingleConnectionPicker(amqpURL string) (*SingleConnectionPicker, error) {
	var conn, err = amqp.Dial(amqpURL)

	if err != nil {
		return nil, err
	}

	picker := &SingleConnectionPicker{
		amqpURL: amqpURL,
		conn:    conn,
		cond:    sync.NewCond(&sync.Mutex{}),
	}

	go picker.supervise()

	return picker, nil
}

func (p *SingleConnectionPicker) supervise() {
	for {
		var (
			ch = make(chan *amqp.Error)

			err error
		)

		p.conn.NotifyClose(ch)
		amqpErr := <-ch

		if amqpErr != nil {
			log.Errorf("AMQPChannelPool connection closed: %v", amqpErr)
		}

		p.cond.L.Lock()
		p.closed = true
		p.cond.L.Unlock()

		p.conn, err = amqp.Dial(p.amqpURL)

		for err != nil {
			log.Errorf("amqp.Dial: %s", err)
			p.conn, err = amqp.Dial(p.amqpURL)
		}

		log.Noticef("Reconnection to RabbitMQ")
		p.closed = false
		p.cond.Broadcast()
	}
}

func (p *SingleConnectionPicker) Pick(ctx context.Context) (*amqp.Connection, func(), error) {
	ch := make(chan *amqp.Connection)

	go func() {
		p.cond.L.Lock()
		defer p.cond.L.Unlock()

		for p.closed {
			p.cond.Wait()
		}

		ch <- p.conn
	}()

	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case res := <-ch:
		return res, func() {}, nil
	}
}
