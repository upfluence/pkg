package thrift

type TPoolProvider interface {
	BuildClient() TClientProvider
	BuildPool(func() (interface{}, error)) (TPool, error)
}

type TPool interface {
	Get(Context) (interface{}, error)
	Put(Context, interface{}) error
	Discard(Context, interface{}) error
}
