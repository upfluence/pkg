package serializer

import (
	"bytes"
	"io"
	"strings"

	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/compress"
	"github.com/upfluence/pkg/thrift/thriftutil"
)

type TDeserializer struct {
	pf thrift.TProtocolFactory

	rfn func(io.Reader) (thrift.TTransport, error)

	cType string
}

func defaultRTBuidler(c compress.Compressor) func(io.Reader) (thrift.TTransport, error) {
	return func(r io.Reader) (thrift.TTransport, error) {
		r, err := c.WrapReader(r)

		if err != nil {
			return nil, err
		}

		return thriftutil.WrapReader(r), nil
	}
}

func NewTDeserializer(pf thrift.TProtocolFactory, cs ...compress.Compressor) *TDeserializer {
	c := compress.CombineCompressors(cs...)

	return &TDeserializer{
		pf:    pf,
		rfn:   defaultRTBuidler(c),
		cType: thriftutil.ExtractContentType(pf, c),
	}
}

func NewDefaultTDeserializer() *TDeserializer {
	return NewTDeserializer(thriftutil.JSONProtocolFactory)
}

func (d *TDeserializer) ContentType() string { return d.cType }

func (d *TDeserializer) ReadFrom(msg TStruct, r io.Reader) error {
	var t, err = d.rfn(r)

	if err != nil {
		return err
	}

	return msg.Read(d.pf.GetProtocol(t))
}

func (d *TDeserializer) Read(msg TStruct, p []byte) error {
	return d.ReadFrom(msg, bytes.NewReader(p))
}

func (d *TDeserializer) ReadString(msg TStruct, p string) error {
	return d.ReadFrom(msg, strings.NewReader(p))
}
