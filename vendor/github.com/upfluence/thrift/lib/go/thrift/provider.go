package thrift

type TClientProvider interface {
	Build(string, string) (TClient, error)
}

type TProcessorProvider interface {
	Build(string, string) (TProcessor, error)
}
