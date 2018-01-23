package thrift

type TClientProvider interface {
	Build(string, string) (TClient, error)
}
