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

type TNopMiddleware struct{}

func (TNopMiddleware) HandleBinaryRequest(ctx Context, _ string, _ int32, req TRequest, next func(Context, TRequest) (TResponse, error)) (TResponse, error) {
	return next(ctx, req)
}

func (TNopMiddleware) HandleUnaryRequest(ctx Context, _ string, _ int32, req TRequest, next func(Context, TRequest) error) error {
	return next(ctx, req)
}

type TMultiMiddleware []TMiddleware

func (ms TMultiMiddleware) HandleBinaryRequest(ctx Context, mth string, seqID int32, req TRequest, next func(Context, TRequest) (TResponse, error)) (TResponse, error) {
	for i := len(ms); i > 0; i-- {
		call := next
		i := i
		next = func(ctx Context, req TRequest) (TResponse, error) {
			return ms[i-1].HandleBinaryRequest(ctx, mth, seqID, req, call)
		}
	}

	return next(ctx, req)
}

func (ms TMultiMiddleware) HandleUnaryRequest(ctx Context, mth string, seqID int32, req TRequest, next func(Context, TRequest) error) error {
	for i := len(ms); i > 0; i-- {
		call := next
		i := i
		next = func(ctx Context, req TRequest) error {
			return ms[i-1].HandleUnaryRequest(ctx, mth, seqID, req, call)
		}
	}

	return next(ctx, req)
}

func WrapMiddlewares(ms []TMiddleware) TMiddleware {
	switch len(ms) {
	case 0:
		return TNopMiddleware{}
	case 1:
		return ms[0]
	}

	return TMultiMiddleware(ms)
}
