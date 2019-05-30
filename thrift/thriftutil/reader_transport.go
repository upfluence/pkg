package thriftutil

import (
	"errors"
	"io"

	"github.com/upfluence/thrift/lib/go/thrift"
)

var (
	ErrReadOnly = errors.New("thriftutil: Transport is read only")

	_ thrift.TTransport = &ReaderTransport{}
)

type ReaderTransport struct {
	io.Reader
	io.Closer
}

func WrapReader(r io.Reader) *ReaderTransport {
	c, ok := r.(io.Closer)

	if !ok {
		c = NopCloser
	}

	return &ReaderTransport{Reader: r, Closer: c}
}

func (r *ReaderTransport) Write(p []byte) (int, error) {
	return 0, ErrReadOnly
}

func (r *ReaderTransport) Flush() error { return ErrReadOnly }
func (r *ReaderTransport) Open() error  { return nil }
func (r *ReaderTransport) IsOpen() bool { return true }

func (r *ReaderTransport) WriteContext(thrift.Context) error {
	return ErrReadOnly
}
