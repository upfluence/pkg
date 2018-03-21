package consumer

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"github.com/upfluence/pkg/log"
)

var ErrCancelled = errors.New("amqp/consumer: Consumer is cancelled")

type Consumer interface {
	Open(context.Context) error
	Consume() (<-chan amqp.Delivery, <-chan *amqp.Error, error)
	QueueName(context.Context) (string, error)
	Close() error
}

type consumer struct {
	opts *options

	queueName string

	cancelFn func()
	openOnce sync.Once

	consumersM *sync.RWMutex
	consumers  []chan amqp.Delivery

	errForwardersM *sync.RWMutex
	errForwarders  []chan *amqp.Error

	closeAck, connectAck chan interface{}
}

func NewConsumer(opts ...Option) Consumer {
	var options = *defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &consumer{
		opts:           &options,
		consumersM:     &sync.RWMutex{},
		errForwardersM: &sync.RWMutex{},
		closeAck:       make(chan interface{}),
		connectAck:     make(chan interface{}),
	}
}

func (c *consumer) QueueName(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-c.connectAck:
		return c.queueName, nil
	}
}

func (c *consumer) loop(ctx context.Context) {
	var i int

	defer close(c.closeAck)

	for {
		var ok, err = c.consume(ctx)

		if ok || !c.opts.shouldContinueFn(err) {
			return
		}

		log.Error(err)

		t, _ := c.opts.backoff.Backoff(i)
		i++

		time.Sleep(t)
	}
}

func (c *consumer) consume(ctx context.Context) (bool, error) {
	select {
	case <-ctx.Done():
		return true, ctx.Err()
	default:
	}

	ch, err := c.opts.pool.Get(ctx)

	if err != nil {
		return false, errors.Wrap(err, "pool.Get")
	}

	if qName := c.opts.queueName; qName == "" {
		q, errQ := ch.QueueDeclare("", false, true, true, false, nil)

		if errQ != nil {
			return false, errors.Wrap(errQ, "channel.QueueDeclare")
		}

		log.Noticef("Queue declared: %s", q.Name)
		c.queueName = q.Name
	} else {
		c.queueName = qName
	}

	ds, err := ch.Consume(
		c.queueName,
		c.opts.consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return false, errors.Wrap(err, "channel.Consume")
	}

	close(c.connectAck)
	defer func() { c.connectAck = make(chan interface{}) }()

	closeCh := make(chan *amqp.Error)
	ch.NotifyClose(closeCh)

	for {
		select {
		case <-ctx.Done():
			ch.Cancel(c.opts.consumerTag, false)
			c.opts.pool.Put(ch)

			return true, ctx.Err()
		case err := <-closeCh:
			for _, f := range c.errForwarders {
				f <- err
			}

			c.opts.pool.Discard(ch)
			return false, nil
		case d := <-ds:
			for _, c := range c.consumers {
				c <- d
			}
		}
	}
}

func (c *consumer) open(ctx context.Context) error {
	var cctx, fn = context.WithCancel(context.Background())
	c.cancelFn = fn

	if err := c.opts.pool.Open(ctx); err != nil {
		return err
	}

	go c.loop(cctx)

	return nil
}

func (c *consumer) Open(ctx context.Context) error {
	var err error

	c.openOnce.Do(func() { err = c.open(ctx) })

	return err
}

func (c *consumer) Consume() (<-chan amqp.Delivery, <-chan *amqp.Error, error) {
	select {
	case _, ok := <-c.closeAck:
		if !ok {
			return nil, nil, ErrCancelled
		}
	default:
	}

	var (
		ch   = make(chan amqp.Delivery)
		errF = make(chan *amqp.Error)
	)

	c.consumersM.Lock()
	defer c.consumersM.Unlock()

	c.consumers = append(c.consumers, ch)

	c.errForwardersM.Lock()
	defer c.errForwardersM.Unlock()

	c.errForwarders = append(c.errForwarders, errF)

	return ch, errF, nil
}

func (c *consumer) Close() error {
	if fn := c.cancelFn; fn != nil {
		fn()

		<-c.closeAck
	}

	for _, c := range c.consumers {
		close(c)
	}

	for _, f := range c.errForwarders {
		close(f)
	}

	return nil
}
