package serializer

import (
	"bytes"
	"io"
	"strings"

	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/encoding"
	"github.com/upfluence/pkg/thrift/thriftutil"
)

type Deserializer interface {
	ContentType() string
	ReadFrom(msg TStruct, r io.Reader) error
	WrapEncoding(encoding.Encoding) Deserializer
}

type tDeserializer struct {
	pf thrift.TProtocolFactory

	rfn func(io.Reader) (thrift.TTransport, error)

	cType string
}

type decoderWrapper struct {
	efn encoding.DecoderFunc
	rfn func(io.Reader) (thrift.TTransport, error)
}

func (dw decoderWrapper) transport(r io.Reader) (thrift.TTransport, error) {
	var err error

	r, err = dw.efn(r)

	if err != nil {
		return nil, err
	}

	return dw.rfn(r)
}

func (td *tDeserializer) ContentType() string { return td.cType }
func (td *tDeserializer) ReadFrom(msg TStruct, r io.Reader) error {
	var t, err = td.rfn(r)

	if err != nil {
		return err
	}

	return msg.Read(td.pf.GetProtocol(t))
}

func (td *tDeserializer) WrapEncoding(e encoding.Encoding) Deserializer {
	return &tDeserializer{
		pf:    td.pf,
		rfn:   decoderWrapper{efn: e.WrapReader, rfn: td.rfn}.transport,
		cType: thriftutil.WrapContentType(td.cType, e),
	}
}

type TDeserializer struct {
	Deserializer
}

func NewTDeserializer(pf thrift.TProtocolFactory, es ...encoding.Encoding) *TDeserializer {
	var td Deserializer = &tDeserializer{
		pf: pf,
		rfn: func(r io.Reader) (thrift.TTransport, error) {
			return thriftutil.WrapReader(r), nil
		},
		cType: thriftutil.ProtocolFactoryContentType(pf),
	}

	for _, e := range es {
		td = td.WrapEncoding(e)
	}

	return &TDeserializer{td}
}

func NewDefaultTDeserializer() *TDeserializer {
	return NewTDeserializer(thriftutil.JSONProtocolFactory)
}

func (d *TDeserializer) Read(msg TStruct, p []byte) error {
	return d.ReadFrom(msg, bytes.NewReader(p))
}

func (d *TDeserializer) ReadString(msg TStruct, p string) error {
	return d.ReadFrom(msg, strings.NewReader(p))
}
