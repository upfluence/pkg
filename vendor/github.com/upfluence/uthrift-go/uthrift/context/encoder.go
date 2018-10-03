package context

import (
	"context"
	"time"

	"github.com/upfluence/pkg/thrift/serializer"
	"github.com/upfluence/thrift/lib/go/thrift"

	tcontext "github.com/upfluence/uthrift-go/uthrift/context/.gen/context"
)

var DefaultEncoder = &Encoder{
	serializer: serializer.NewDefaultTSerializerFactory(),
}

type Encoder struct {
	serializer *serializer.TSerializerFactory
}

func (e *Encoder) Decode(buf []byte, initialCtx thrift.Context) (thrift.Context, func(), error) {
	var (
		res = &Context{
			Context: initialCtx,
			payload: &tcontext.Context{},
		}
		cancel = func() {}
	)

	if len(buf) == 0 {
		return res, cancel, nil
	}

	if err := e.serializer.GetSerializer().Read(res.payload, buf); err != nil {
		return nil, cancel, err
	}

	res.Context = context.WithValue(res.Context, thriftCtxKey, res.payload.Values)

	if dSec := res.payload.DeadlineSec; dSec != nil {
		dNano := res.payload.GetDeadlineNano()

		res.Context, cancel = context.WithDeadline(res.Context, time.Unix(*dSec, dNano))
	}

	return res, cancel, nil
}

func (e *Encoder) Encode(ctx thrift.Context) ([]byte, error) {
	return e.serializer.GetSerializer().Write(transform(ctx).payload)
}

func Encode(ctx thrift.Context) ([]byte, error) {
	return DefaultEncoder.Encode(ctx)
}

func Decode(buf []byte, initialCtx thrift.Context) (thrift.Context, func(), error) {
	return DefaultEncoder.Decode(buf, initialCtx)
}
