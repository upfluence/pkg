package thriftutil

import (
	"errors"
	"io"

	"github.com/upfluence/thrift/lib/go/thrift"
)

var (
	NopFlusher = nopFlusher{}
	NopCloser  = nopCloser{}

	ErrWriteOnly = errors.New("thriftutil: Transport is write only")

	_ thrift.TTransport = &WriterTransport{}
)

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

type nopFlusher struct{}

func (nopFlusher) Flush() error { return nil }

type WriterTransport struct {
	io.Writer
	io.Closer
	thrift.Flusher
}

func (*WriterTransport) WriteContext(thrift.Context) error { return nil }
func (*WriterTransport) Open() error                       { return nil }
func (*WriterTransport) IsOpen() bool                      { return true }

func (*WriterTransport) Read([]byte) (int, error) {
	return 0, ErrWriteOnly
}

func WrapWriter(w io.Writer) *WriterTransport {
	c, ok := w.(io.Closer)

	if !ok {
		c = NopCloser
	}

	f, ok := w.(thrift.Flusher)

	if !ok {
		f = NopFlusher
	}

	return &WriterTransport{Writer: w, Closer: c, Flusher: f}
}
