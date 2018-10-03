package clientprovider

import (
	"fmt"

	"github.com/upfluence/pkg/pool/bounded"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/uthrift-go/uthrift/client/pool"
	"github.com/upfluence/uthrift-go/uthrift/middleware"
	"github.com/upfluence/uthrift-go/uthrift/middleware/logger"
	"github.com/upfluence/uthrift-go/uthrift/middleware/prometheus"
	"github.com/upfluence/uthrift-go/uthrift/middleware/stapler"
)

const (
	defaultPoolSize = 5

	client    component = "client"
	protocol  component = "protocol"
	transport component = "transport"
)

var (
	initializeBare = func() *Provider {
		return &Provider{
			clientFactory:   thrift.NewTDefaultClientFactory(),
			protocolFactory: thrift.NewTBinaryProtocolFactoryDefault(),
		}
	}

	initializeDefault = func() *Provider {
		return &Provider{
			clientFactory:   pool.NewFactory(bounded.NewPoolFactory(defaultPoolSize)),
			protocolFactory: thrift.NewTBinaryProtocolFactoryDefault(),
			middlewareFactories: []middleware.Factory{
				logger.NewFactory(),
				prometheus.NewFactory(prometheus.Client),
				stapler.NewDefaultFactory(),
			},
		}
	}
)

type component string

type MissingComponentError struct {
	c component
}

func (e *MissingComponentError) Error() string {
	return fmt.Sprintf(
		"Missing %s factory in the provider to create a  functional client",
		e.c,
	)
}

type ProviderOptions func(*Provider)

type Provider struct {
	clientFactory       thrift.TClientFactory
	protocolFactory     thrift.TProtocolFactory
	transportFactory    thrift.TTransportFactory
	middlewareFactories []middleware.Factory
}

func newProvider(init func() *Provider, opts []ProviderOptions) *Provider {
	var p = init()

	p.Apply(opts...)

	return p
}

func NewBareProvider(opts ...ProviderOptions) *Provider {
	return newProvider(initializeBare, opts)
}

func NewDefaultProvider(opts ...ProviderOptions) *Provider {
	return newProvider(initializeDefault, opts)
}

func (p *Provider) Apply(opts ...ProviderOptions) {
	for _, opt := range opts {
		opt(p)
	}
}

func (p *Provider) Build(ns, svc string) (thrift.TClient, error) {
	if p.clientFactory == nil {
		return nil, &MissingComponentError{client}
	}

	if p.protocolFactory == nil {
		return nil, &MissingComponentError{protocol}
	}

	if p.transportFactory == nil {
		return nil, &MissingComponentError{transport}
	}

	var middlewares = make([]thrift.TMiddleware, len(p.middlewareFactories))

	for i, f := range p.middlewareFactories {
		middlewares[i] = f.GetMiddleware(ns, svc)
	}

	return p.clientFactory.GetClient(
		p.transportFactory,
		p.protocolFactory,
		middlewares,
	), nil
}
