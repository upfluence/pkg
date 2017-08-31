package thrift

type TClientProvider interface {
	Build(string) (TTransport, TProtocolFactory, error)
}

type TServerProvider interface {
	Build(string) (TServerFactory, TProtocolFactory, error)
}
