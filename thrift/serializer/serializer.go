package serializer

import (
	"io"

	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/bytesutil"
	"github.com/upfluence/pkg/encoding"
	"github.com/upfluence/pkg/thrift/thriftutil"
)

type Serializer interface {
	ContentType() string
	WriteTo(TStruct, io.Writer) error
	WrapEncoding(encoding.Encoding) Serializer
}

type tSerializer struct {
	pf thrift.TProtocolFactory

	wfn func(io.Writer) (thrift.TTransport, error)

	cType string
}

type encoderWrapper struct {
	efn encoding.EncoderFunc
	wfn func(io.Writer) (thrift.TTransport, error)
}

func (ew encoderWrapper) transport(w io.Writer) (thrift.TTransport, error) {
	var err error
	w, err = ew.efn(w)

	if err != nil {
		return nil, err
	}

	return ew.wfn(w)
}

func (ts *tSerializer) ContentType() string { return ts.cType }

func (ts *tSerializer) WrapEncoding(e encoding.Encoding) Serializer {
	return &tSerializer{
		pf:    ts.pf,
		wfn:   encoderWrapper{efn: e.WrapWriter, wfn: ts.wfn}.transport,
		cType: thriftutil.WrapContentType(ts.cType, e),
	}
}

func (ts *tSerializer) WriteTo(msg TStruct, w io.Writer) error {
	var t, err = ts.wfn(w)

	if err != nil {
		return err
	}

	p := ts.pf.GetProtocol(t)

	if err := msg.Write(p); err != nil {
		return err
	}

	if err := p.Flush(); err != nil {
		return err
	}

	if err := t.Flush(); err != nil {
		return err
	}

	return t.Close()
}

type TSerializer struct {
	Serializer
	bp *bytesutil.BufferPool
}

func NewTSerializer(pf thrift.TProtocolFactory, cs ...encoding.Encoding) *TSerializer {
	var ts Serializer = &tSerializer{
		pf: pf,
		wfn: func(w io.Writer) (thrift.TTransport, error) {
			return thriftutil.WrapWriter(w), nil
		},
		cType: thriftutil.ProtocolFactoryContentType(pf),
	}

	for _, c := range cs {
		ts = ts.WrapEncoding(c)
	}

	return &TSerializer{Serializer: ts, bp: bytesutil.NewBufferPool()}
}

func NewDefaultTSerializer() *TSerializer {
	return NewTSerializer(thriftutil.JSONProtocolFactory)
}

func (s *TSerializer) Write(msg TStruct) ([]byte, error) {
	buf := s.bp.Get()

	err := s.WriteTo(msg, buf)
	res := make([]byte, buf.Len())
	copy(res, buf.Bytes())

	s.bp.Put(buf)

	return res, err
}

func (s *TSerializer) WriteString(msg TStruct) (string, error) {
	buf := s.bp.Get()

	err := s.WriteTo(msg, buf)
	p := buf.String()

	s.bp.Put(buf)

	return p, err
}
