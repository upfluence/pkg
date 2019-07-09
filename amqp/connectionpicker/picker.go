package connectionpicker

import (
	"context"
	"sync"

	"github.com/streadway/amqp"

	"github.com/upfluence/pkg/amqp/util"
	"github.com/upfluence/pkg/closer"
	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/group"
	"github.com/upfluence/pkg/log"
	"github.com/upfluence/pkg/multierror"
)

type Picker interface {
	closer.Shutdowner

	Open(context.Context) error
	IsOpen() bool

	Pick(context.Context) (*amqp.Connection, error)
}

type picker struct {
	*options
	*closer.Monitor

	sync.Mutex
	cs []*amqp.Connection

	lastUsedConn int
}

func NewPicker(opts ...Option) Picker {
	options := *defaultOptions

	for _, opt := range opts {
		opt(&options)
	}

	return &picker{options: &options, Monitor: closer.NewMonitor()}
}

func (p *picker) IsOpen() bool {
	return p.Balancer.IsOpen() && p.Monitor.IsOpen()
}

func (p *picker) Shutdown(ctx context.Context) error {
	g := group.WaitGroup(ctx)

	if sb, ok := p.Balancer.(closer.Shutdowner); ok {
		g.Do(sb.Shutdown)
	}

	g.Do(p.Monitor.Shutdown)

	return g.Wait()
}

func (p *picker) Close() error {
	return multierror.Combine(
		p.Balancer.Close(),
		p.Monitor.Close(),
	)
}

func (p *picker) Pick(ctx context.Context) (*amqp.Connection, error) {
	p.Lock()
	defer p.Unlock()

	if len(p.cs) < p.targetOpenedConn {
		peer, err := p.Balancer.Get(ctx, balancer.BalancerGetOptions{})

		if err != nil {
			return nil, err
		}

		conn, err := util.Dial(
			ctx,
			peer,
			p.peer,
			p.connectionNamer(len(p.cs)),
		)

		if conn != nil {
			p.Run(func(ctx context.Context) { p.supervise(ctx, conn) })
			p.cs = append(p.cs, conn)
		}

		return conn, err
	}

	p.lastUsedConn = (p.lastUsedConn + 1) % p.targetOpenedConn

	return p.cs[p.lastUsedConn], nil
}

func (p *picker) supervise(ctx context.Context, conn *amqp.Connection) {
	var ch = make(chan *amqp.Error)

	conn.NotifyClose(ch)

	select {
	case err := <-ch:
		if err != nil {
			log.WithError(err).Warning("Connection closed")
		}
	case <-ctx.Done():
		conn.Close()

	}
	p.Lock()
	var cs []*amqp.Connection

	for _, wConn := range p.cs {
		if wConn != conn {
			cs = append(cs, wConn)
		}
	}

	p.cs = cs
	p.Unlock()
}
