package serializer

import (
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/v2/encoding"
	"github.com/upfluence/pkg/v2/thrift/thriftutil"
)

type TSerializerFactory struct {
	s *TSerializer
	d *TDeserializer
}

func NewDefaultTSerializerFactory() *TSerializerFactory {
	return NewTSerializerFactory(thriftutil.JSONProtocolFactory)
}

func NewTSerializerFactory(pf thrift.TProtocolFactory, es ...encoding.Encoding) *TSerializerFactory {
	return &TSerializerFactory{
		s: NewTSerializer(pf, es...),
		d: NewTDeserializer(pf, es...),
	}
}

func (sf *TSerializerFactory) GetSerializer() *TSerializer     { return sf.s }
func (sf *TSerializerFactory) GetDeserializer() *TDeserializer { return sf.d }
