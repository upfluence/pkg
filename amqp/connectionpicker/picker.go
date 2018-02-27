package connectionpicker

import (
	"context"
	"sync"

	"github.com/streadway/amqp"

	"github.com/upfluence/pkg/amqp/util"
	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/log"
)

type Picker interface {
	Open(context.Context) error
	IsOpen() bool
	Close() error

	Pick(context.Context) (*amqp.Connection, error)
}

type picker struct {
	*options

	watchedConnections   []*amqp.Connection
	watchedConnectionsMu *sync.Mutex

	lastUsedConn int
}

func NewPicker(opts ...Option) Picker {
	options := *defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &picker{
		options:              &options,
		watchedConnectionsMu: &sync.Mutex{},
	}
}

func (p *picker) Close() error {
	p.watchedConnectionsMu.Lock()
	defer p.watchedConnectionsMu.Unlock()

	for _, conn := range p.watchedConnections {
		conn.Close()
	}

	return p.options.Close()
}

func (p *picker) Pick(ctx context.Context) (*amqp.Connection, error) {
	p.watchedConnectionsMu.Lock()
	defer p.watchedConnectionsMu.Unlock()

	if len(p.watchedConnections) < p.targetOpenedConn {
		peer, err := p.Balancer.Get(ctx, balancer.BalancerGetOptions{})

		if err != nil {
			return nil, err
		}

		conn, err := util.Dial(
			ctx,
			peer,
			p.peer,
			p.connectionNamer(len(p.watchedConnections)),
		)

		log.Noticef("conn: %v, err: %v", conn, err)

		if conn != nil {
			go p.supervise(conn)
			p.watchedConnections = append(p.watchedConnections, conn)
		}

		return conn, err
	}

	p.lastUsedConn = (p.lastUsedConn + 1) % p.targetOpenedConn

	return p.watchedConnections[p.lastUsedConn], nil
}

func (p *picker) supervise(conn *amqp.Connection) {
	var ch = make(chan *amqp.Error)

	conn.NotifyClose(ch)

	err := <-ch

	if err != nil {
		log.Errorf("Connection closed: %v", err)
	}

	p.watchedConnectionsMu.Lock()
	defer p.watchedConnectionsMu.Unlock()

	var watchedConns = []*amqp.Connection{}

	for _, wConn := range p.watchedConnections {
		if wConn != conn {
			watchedConns = append(watchedConns, wConn)
		}
	}

	p.watchedConnections = watchedConns
}
