package clientprovider

import (
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/uthrift-go/uthrift/middleware"
)

func WithTransport(f thrift.TTransportFactory) ProviderOptions {
	return func(p *Provider) {
		p.transportFactory = f
	}
}

func WithProtocol(f thrift.TProtocolFactory) ProviderOptions {
	return func(p *Provider) {
		p.protocolFactory = f
	}
}

func WithClient(f thrift.TClientFactory) ProviderOptions {
	return func(p *Provider) {
		p.clientFactory = f
	}
}

func WithMiddlewares(fs ...middleware.Factory) ProviderOptions {
	return func(p *Provider) {
		for _, f := range fs {
			p.middlewareFactories = append(p.middlewareFactories, f)
		}
	}
}
