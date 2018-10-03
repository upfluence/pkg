package base

import (
	"bytes"
	stdcontext "context"

	"github.com/upfluence/thrift/lib/go/thrift"
)

type Transport struct {
	Ctx thrift.Context
	Buf *bytes.Buffer
}

func NewTransport() *Transport {
	return &Transport{
		Ctx: stdcontext.Background(),
		Buf: &bytes.Buffer{},
	}
}

func (t *Transport) Write(data []byte) (int, error) {
	n, err := t.Buf.Write(data)

	return n, thrift.NewTTransportExceptionFromError(err)
}

func (t *Transport) WriteContext(ctx thrift.Context) error {
	t.Ctx = ctx

	return nil
}

func (t *Transport) ResetBuffers() {
	t.Ctx = stdcontext.Background()
	t.Buf.Reset()
}
