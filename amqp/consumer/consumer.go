package consumer

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"github.com/upfluence/pkg/closer"
	"github.com/upfluence/pkg/log"
)

var ErrCancelled = errors.New("amqp/consumer: Consumer is cancelled")

type Consumer interface {
	Open(context.Context) error
	IsOpen() bool

	Consume() (<-chan amqp.Delivery, <-chan *amqp.Error, error)
	QueueName(context.Context) (string, error)

	closer.Closer
}

type consumer struct {
	*closer.Monitor

	opts *options

	openOnce sync.Once

	consumersM sync.Mutex
	consumers  []chan amqp.Delivery

	errForwardersM sync.Mutex
	errForwarders  []chan *amqp.Error

	queueCond *sync.Cond
	queueM    sync.Mutex
	queue     string
}

func NewConsumer(opts ...Option) Consumer {
	var options = *defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	c := consumer{Monitor: closer.NewMonitor(), opts: &options}
	c.queueCond = sync.NewCond(&c.queueM)

	return &c
}

func (c *consumer) QueueName(ctx context.Context) (string, error) {
	c.queueM.Lock()
	q := c.queue
	c.queueM.Unlock()

	if q != "" {
		return q, nil
	}

	done := make(chan struct{})
	cancelled := false

	go func() {
		c.queueM.Lock()
		defer c.queueM.Unlock()

		for {
			if cancelled || c.queue != "" {
				q = c.queue
				close(done)
				return
			}

			c.queueCond.Wait()
		}
	}()

	select {
	case <-ctx.Done():
		c.queueM.Lock()
		cancelled = true
		c.queueCond.Broadcast()
		c.queueM.Unlock()

		return "", ctx.Err()
	case <-done:
		return q, nil
	}
}

func (c *consumer) loop(ctx context.Context) {
	var i int

	for {
		var ok, err = c.consume(ctx)

		if ok || !c.opts.shouldContinueFn(err) {
			return
		}

		log.WithError(err).Warning("cant consume")

		t, _ := c.opts.backoff.Backoff(i)
		i++

		time.Sleep(t)
	}
}

func (c *consumer) consume(ctx context.Context) (bool, error) {
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
		c.queueM.Lock()
		c.queue = q.Name
		c.queueM.Unlock()
	} else {
		c.queueM.Lock()
		c.queue = qName
		c.queueM.Unlock()
	}

	ds, err := ch.Consume(
		c.queue,
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

	c.queueCond.Broadcast()

	closeCh := make(chan *amqp.Error)
	ch.NotifyClose(closeCh)

	for {
		select {
		case <-ctx.Done():
			c.queueM.Lock()
			c.queue = ""
			c.queueM.Unlock()

			ch.Cancel(c.opts.consumerTag, false)
			c.opts.pool.Put(ch)

			return true, ctx.Err()
		case err := <-closeCh:
			c.queueM.Lock()
			c.queue = ""
			c.queueM.Unlock()

			for _, f := range c.errForwarders {
				select {
				case <-ctx.Done():
				case f <- err:
				}
			}

			c.opts.pool.Discard(ch)
			return false, nil
		case d := <-ds:
			for _, c := range c.consumers {
				select {
				case <-ctx.Done():
				case c <- d:
				}
			}

			d.Ack(false)
		}
	}
}

func (c *consumer) open(ctx context.Context) error {
	if err := c.opts.pool.Open(ctx); err != nil {
		return err
	}

	c.Run(c.loop)

	return nil
}

func (c *consumer) Open(ctx context.Context) error {
	var err error

	c.openOnce.Do(func() { err = c.open(ctx) })

	return err
}

func (c *consumer) Consume() (<-chan amqp.Delivery, <-chan *amqp.Error, error) {
	if !c.IsOpen() {
		return nil, nil, ErrCancelled
	}

	var (
		ch   = make(chan amqp.Delivery)
		errF = make(chan *amqp.Error)
	)

	c.consumersM.Lock()
	c.consumers = append(c.consumers, ch)
	c.consumersM.Unlock()

	c.errForwardersM.Lock()
	c.errForwarders = append(c.errForwarders, errF)
	c.errForwardersM.Unlock()

	return ch, errF, nil
}

func (c *consumer) Close() error {
	c.Monitor.Close()

	c.consumersM.Lock()
	for _, c := range c.consumers {
		close(c)
	}
	c.consumers = nil
	c.consumersM.Unlock()

	c.errForwardersM.Lock()
	for _, f := range c.errForwarders {
		close(f)
	}
	c.errForwarders = nil
	c.errForwardersM.Unlock()

	if c.opts.handlePoolClosing {
		return c.opts.pool.Close()
	}

	return nil
}
