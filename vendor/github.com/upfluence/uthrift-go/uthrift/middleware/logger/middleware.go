package logger

import (
	"fmt"
	"time"

	"github.com/upfluence/pkg/log"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/uthrift-go/pkg/responseutil"
)

const success = "success"

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (*Factory) GetMiddleware(ns, svc string) thrift.TMiddleware {
	return NewMiddleware(ns, svc)
}

type Middleware struct {
	namespace, service string
}

func NewMiddleware(namespace, service string) *Middleware {
	return &Middleware{namespace: namespace, service: service}
}

func (m *Middleware) HandleBinaryRequest(ctx thrift.Context, mth string, seqID int32, req thrift.TRequest, next func(thrift.Context, thrift.TRequest) (thrift.TResponse, error)) (thrift.TResponse, error) {
	var (
		err error
		res thrift.TResponse
	)

	m.handle(
		mth,
		func() error {
			res, err = next(ctx, req)

			return responseutil.FetchError(res, err)
		},
	)

	return res, err

}

func (m *Middleware) HandleUnaryRequest(ctx thrift.Context, mth string, seqID int32, req thrift.TRequest, next func(thrift.Context, thrift.TRequest) error) error {
	var err error

	m.handle(
		mth,
		func() error {
			err = next(ctx, req)

			return err
		},
	)

	return err
}

func (m *Middleware) handle(mth string, next func() error) {
	var (
		t0     = time.Now()
		status = success
		err    = next()
	)

	if err != nil {
		status = fmt.Sprintf("%T: %s", err, err.Error())
	}

	log.Noticef(
		"%s.%s#%s: %v: %s",
		m.namespace,
		m.service,
		mth,
		time.Since(t0),
		status,
	)
}
