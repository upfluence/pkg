package thriftutil

import (
	"fmt"
	"strings"

	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/compress"
)

var (
	BinaryProtocolFactory = WithContentType(thrift.NewTBinaryProtocolFactoryDefault(), "application/binary")
	JSONProtocolFactory   = WithContentType(thrift.NewTSimpleJSONProtocolFactory(), "application/binary")
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

func pfContentType(pf thrift.TProtocolFactory) string {
	if ct, ok := pf.(ContentTyper); ok {
		return ct.ContentType()
	}

	switch pf.(type) {
	case *thrift.TBinaryProtocolFactory:
		return "application/binary"
	case *thrift.TSimpleJSONProtocolFactory, *thrift.TJSONProtocolFactory:
		return "application/json"
	}

	return fmt.Sprintf("application/%T", pf)
}

func ExtractContentType(pf thrift.TProtocolFactory, c compress.Compressor) string {
	var fs = []string{pfContentType(pf)}

	if ct := c.ContentType(); ct != "" {
		fs = append(fs, ct)
	}

	return strings.Join(fs, "+")
}
