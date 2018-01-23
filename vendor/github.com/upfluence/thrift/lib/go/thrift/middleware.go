package thrift

type TResponse interface {
	TStruct
	GetResult() interface{}
	GetError() error
}

type TRequest interface {
	TStruct
}

type TMiddleware interface {
	HandleBinaryRequest(ctx Context, mth string, seqID int32, req TRequest, next func(Context, TRequest) (TResponse, error)) (TResponse, error)
	HandleUnaryRequest(ctx Context, mth string, seqID int32, req TRequest, next func(Context, TRequest) error) error
}

type TMiddlewareBuilder interface {
	Build(namespace, service string) TMiddleware
}
