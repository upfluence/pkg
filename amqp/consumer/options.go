package consumer

import (
	"time"

	"github.com/upfluence/pkg/amqp/channelpool"
	"github.com/upfluence/pkg/amqp/connectionpicker"
	"github.com/upfluence/pkg/backoff"
	"github.com/upfluence/pkg/backoff/static"
	"github.com/upfluence/pkg/peer"
	"github.com/upfluence/pkg/pool/bounded"
)

var defaultOptions = &options{
	pool: channelpool.NewPool(
		bounded.NewPoolFactory(1),
		connectionpicker.NewPicker(),
	),
	handlePoolClosing: true,
	shouldContinueFn:  func(error) bool { return true },
	backoff:           static.NewInfiniteBackoff(1 * time.Second),
	consumerTag:       peer.FromEnv().InstanceName,
}

type Option func(*options)

type options struct {
	pool              channelpool.Pool
	handlePoolClosing bool

	queueName   string
	consumerTag string

	shouldContinueFn func(error) bool
	backoff          backoff.Strategy
}

func WithPool(p channelpool.Pool) Option {
	return func(o *options) {
		o.pool = p
		o.handlePoolClosing = false
	}
}

func WithQueueName(n string) Option {
	return func(o *options) { o.queueName = n }
}

func WithConsumerTag(t string) Option {
	return func(o *options) { o.consumerTag = t }
}

func WithBackoff(s backoff.Strategy) Option {
	return func(o *options) { o.backoff = s }
}
