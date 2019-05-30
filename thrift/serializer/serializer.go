package serializer

import (
	"io"

	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/bytesutil"
	"github.com/upfluence/pkg/compress"
	"github.com/upfluence/pkg/thrift/thriftutil"
)

type TSerializer struct {
	pf thrift.TProtocolFactory
	bp *bytesutil.BufferPool

	wfn func(io.Writer) (thrift.TTransport, error)

	cType string
}

func defaultWTBuidler(c compress.Compressor) func(io.Writer) (thrift.TTransport, error) {
	return func(w io.Writer) (thrift.TTransport, error) {
		w, err := c.WrapWriter(w)

		if err != nil {
			return nil, err
		}

		return thriftutil.WrapWriter(w), nil
	}
}

func NewTSerializer(pf thrift.TProtocolFactory, cs ...compress.Compressor) *TSerializer {
	c := compress.CombineCompressors(cs...)

	return &TSerializer{
		pf:    pf,
		bp:    bytesutil.NewBufferPool(),
		wfn:   defaultWTBuidler(c),
		cType: thriftutil.ExtractContentType(pf, c),
	}
}

func NewDefaultTSerializer() *TSerializer {
	return NewTSerializer(thriftutil.JSONProtocolFactory)
}

func (s *TSerializer) ContentType() string { return s.cType }

func (s *TSerializer) WriteTo(msg TStruct, w io.Writer) error {
	var t, err = s.wfn(w)

	if err != nil {
		return err
	}

	p := s.pf.GetProtocol(t)

	if err := msg.Write(p); err != nil {
		return err
	}

	return t.Flush()
}

func (s *TSerializer) Write(msg TStruct) ([]byte, error) {
	buf := s.bp.Get()

	err := s.WriteTo(msg, buf)
	p := buf.Bytes()

	s.bp.Put(buf)

	return p, err
}

func (s *TSerializer) WriteString(msg TStruct) (string, error) {
	buf := s.bp.Get()

	err := s.WriteTo(msg, buf)
	p := buf.String()

	s.bp.Put(buf)

	return p, err
}
