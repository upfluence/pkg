package pool

import (
	"context"
	"sync"

	"github.com/upfluence/pkg/log"
	"github.com/upfluence/pkg/pool"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type Client struct {
	pool         pool.Pool
	middlewares  []thrift.TMiddleware
	transportMap *sync.Map
}

type clientFactory struct {
	poolFactory pool.PoolFactory
}

func NewFactory(f pool.PoolFactory) thrift.TClientFactory {
	return &clientFactory{poolFactory: f}
}

func (f *clientFactory) GetClient(t thrift.TTransportFactory, p thrift.TProtocolFactory, ms []thrift.TMiddleware) thrift.TClient {
	return NewClient(f.poolFactory, t, p, ms...)
}

func NewClient(
	poolFactory pool.PoolFactory,
	transportFactory thrift.TTransportFactory,
	protocolFactory thrift.TProtocolFactory,
	ms ...thrift.TMiddleware,
) *Client {
	var transportMap = &sync.Map{}

	return &Client{
		middlewares:  ms,
		transportMap: transportMap,
		pool: poolFactory.GetPool(
			func(context.Context) (interface{}, error) {
				var (
					tr = transportFactory.GetTransport(nil)
					cl = thrift.NewTSyncClient(tr, protocolFactory)
				)

				transportMap.Store(cl, tr)

				return cl, tr.Open()
			},
		),
	}
}

func shouldDiscard(err error) bool {
	switch e := err.(type) {
	case thrift.TApplicationException:
		return e.TypeId() == thrift.BAD_SEQUENCE_ID || e.TypeId() == thrift.PROTOCOL_ERROR
	case thrift.TTransportException:
		return true
	case thrift.TProtocolException:
		return e.TypeId() == thrift.INVALID_DATA || e.TypeId() == thrift.UNKNOWN_PROTOCOL_EXCEPTION
	}

	return false
}

func (c *Client) CallBinary(ctx thrift.Context, method string, req thrift.TRequest, res thrift.TResponse) error {
	var call = func(ctx thrift.Context, req thrift.TRequest) (thrift.TResponse, error) {
		return res, c.callBinary(ctx, method, req, res)
	}

	for i := len(c.middlewares); i > 0; i-- {
		next := call
		j := i - 1
		call = func(ctx thrift.Context, req thrift.TRequest) (thrift.TResponse, error) {
			return c.middlewares[j].HandleBinaryRequest(
				ctx,
				method,
				0,
				req,
				next,
			)
		}
	}

	_, err := call(ctx, req)

	return err
}

func (c *Client) CallUnary(ctx thrift.Context, method string, req thrift.TRequest) error {
	var call = func(ctx thrift.Context, req thrift.TRequest) error {
		return c.callUnary(ctx, method, req)
	}

	for i := len(c.middlewares); i > 0; i-- {
		next := call
		j := i - 1

		call = func(ctx thrift.Context, req thrift.TRequest) error {
			return c.middlewares[j].HandleUnaryRequest(
				ctx,
				method,
				0,
				req,
				next,
			)
		}
	}

	return call(ctx, req)
}

func (c *Client) callBinary(ctx thrift.Context, method string, req thrift.TRequest, res thrift.TResponse) error {
	e, err := c.pool.Get(ctx)

	if err != nil {
		return err
	}

	err = e.(*thrift.TSyncClient).CallBinary(ctx, method, req, res)

	if shouldDiscard(err) {
		if v, ok := c.transportMap.Load(e); ok {
			if err := v.(thrift.TTransport).Close(); err != nil {
				log.Warningf("Close transport: %v", err)
			}
		}

		c.pool.Discard(e)
	} else {
		c.pool.Put(e)
	}

	return err
}

func (c *Client) callUnary(ctx thrift.Context, method string, req thrift.TRequest) error {
	e, err := c.pool.Get(ctx)

	if err != nil {
		return err
	}

	err = e.(*thrift.TSyncClient).CallUnary(ctx, method, req)

	if shouldDiscard(err) {
		if v, ok := c.transportMap.Load(e); ok {
			if err := v.(thrift.TTransport).Close(); err != nil {
				log.Warningf("Close transport: %v", err)
			}
		}

		c.pool.Discard(e)
	} else {
		c.pool.Put(e)
	}

	return err
}
