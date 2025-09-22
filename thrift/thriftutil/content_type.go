package thriftutil

import (
	"fmt"

	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/v2/encoding"
)

var (
	BinaryProtocolFactory = WithContentType(thrift.NewTBinaryProtocolFactoryDefault(), "application/binary")
	JSONProtocolFactory   = WithContentType(thrift.NewTJSONProtocolFactory(), "application/json")
)

type ContentTyper interface {
	ContentType() string
}

type TTypedProtocolFactory interface {
	ContentTyper
	thrift.TProtocolFactory
}

type tTypedProtocolFactory struct {
	thrift.TProtocolFactory
	t string
}

func WithContentType(pf thrift.TProtocolFactory, ct string) TTypedProtocolFactory {
	return &tTypedProtocolFactory{TProtocolFactory: pf, t: ct}
}

func (ttpf *tTypedProtocolFactory) ContentType() string { return ttpf.t }

func ProtocolFactoryContentType(pf thrift.TProtocolFactory) string {
	if ct, ok := pf.(ContentTyper); ok {
		return ct.ContentType()
	}

	switch pf.(type) {
	case *thrift.TBinaryProtocolFactory:
		return "application/binary"
	case *thrift.TJSONProtocolFactory:
		return "application/json"
	}

	return fmt.Sprintf("application/%T", pf)
}

func WrapContentType(ct string, c encoding.Encoding) string {
	return fmt.Sprintf("%s+%s", ct, c.ContentType())
}

func ExtractContentType(pf thrift.TProtocolFactory, c encoding.Encoding) string {
	return WrapContentType(ProtocolFactoryContentType(pf), c)
}
