package context

import (
	"context"

	"github.com/upfluence/pkg/log"
	"github.com/upfluence/pkg/peer"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/base/version"
	tcontext "github.com/upfluence/uthrift-go/uthrift/context/.gen/context"
)

var thriftCtxKey ctxKey = "thriftCtxKey"

type ctxKey string

type Context struct {
	context.Context

	payload *tcontext.Context
}

func transform(ctx thrift.Context) *Context {
	if cctx, ok := ctx.(*Context); ok {
		return cctx
	}

	var payload = &tcontext.Context{}

	if d, ok := ctx.Deadline(); ok {
		payload.DeadlineSec = thrift.Int64Ptr(d.Unix())
		payload.DeadlineNano = thrift.Int64Ptr(d.Unix())
	}

	if value := ctx.Value(thriftCtxKey); value != nil {
		switch v := value.(type) {
		case map[string][]byte:
			payload.Values = v
		default:
			log.Debugf("Cannot parse context value %+v (%T)", v, v)
		}
	} else {
		log.Debug("Context value empty")
	}

	return &Context{Context: ctx, payload: payload}
}

func WithPeer(parent thrift.Context, p *peer.Peer) thrift.Context {
	var (
		versions = make(map[string]*version.Version)
		ctx      = transform(parent)

		stringPtr = func(s string) *string {
			if s == "" {
				return nil
			}

			return &s
		}
	)

	for _, i := range p.Interfaces {
		versions[i.Name()] = i.Version()
	}

	ctx.payload.Client = &tcontext.Peer{
		InstanceName:      stringPtr(p.InstanceName),
		AppName:           stringPtr(p.AppName),
		ProjectName:       stringPtr(p.ProjectName),
		Environment:       stringPtr(p.Environment),
		Version:           p.Version,
		InterfaceVersions: versions,
	}

	return ctx
}

func WithSpanID(parent thrift.Context, spanID string) thrift.Context {
	var ctx = transform(parent)

	ctx.payload.SpanID = thrift.StringPtr(spanID)

	if id := ctx.payload.TraceID; id == nil || *id == "" {
		ctx.payload.TraceID = thrift.StringPtr(spanID)
	}

	return ctx
}
