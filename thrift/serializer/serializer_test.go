package serializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/errors/errtest"
	"github.com/upfluence/pkg/v2/encoding"
)

func TestContentType(t *testing.T) {
	for _, tt := range []struct {
		pf  thrift.TProtocolFactory
		es  []encoding.Encoding
		out string
	}{
		{
			pf:  thrift.NewTJSONProtocolFactory(),
			out: "application/json",
		},
		{
			pf:  thrift.NewTJSONProtocolFactory(),
			es:  []encoding.Encoding{encoding.Base64Encoding},
			out: "application/json+base64",
		},
		{
			pf: thrift.NewTBinaryProtocolFactoryDefault(),
			es: []encoding.Encoding{
				encoding.GZipEncoding,
				encoding.Base64Encoding,
			},
			out: "application/binary+gzip+base64",
		},
	} {
		s := NewTSerializer(tt.pf, tt.es...)
		d := NewTDeserializer(tt.pf, tt.es...)

		assert.Equal(t, tt.out, s.ContentType())
		assert.Equal(t, tt.out, d.ContentType())
	}
}

type stringTStruct struct {
	string
}

func (sts *stringTStruct) Write(p thrift.TProtocol) error {
	return p.WriteString(sts.string)
}

func (sts *stringTStruct) Read(p thrift.TProtocol) error {
	s, err := p.ReadString()

	if err != nil {
		return err
	}

	sts.string = s

	return nil
}

func TestSerializerWriteString(t *testing.T) {
	for _, tt := range []struct {
		name  string
		pf    thrift.TProtocolFactory
		es    []encoding.Encoding
		in    TStruct
		out   string
		errfn errtest.ErrorAssertion
	}{
		{
			name:  "regular binary",
			pf:    thrift.NewTBinaryProtocolFactoryDefault(),
			in:    &stringTStruct{"foobar"},
			out:   "\x00\x00\x00\x06foobar",
			errfn: errtest.NoError(),
		},
		{
			name:  "regular encoding",
			pf:    thrift.NewTJSONProtocolFactory(),
			in:    &stringTStruct{"foobar"},
			out:   "\"foobar\"",
			errfn: errtest.NoError(),
		},
		{
			name:  "base64",
			pf:    thrift.NewTJSONProtocolFactory(),
			es:    []encoding.Encoding{encoding.Base64Encoding},
			in:    &stringTStruct{"foobar"},
			out:   "ImZvb2JhciI=",
			errfn: errtest.NoError(),
		},
		{
			name: "gzip",
			pf:   thrift.NewTJSONProtocolFactory(),
			es: []encoding.Encoding{
				encoding.GZipEncoding,
			},
			in:    &stringTStruct{"foobar"},
			out:   "\x1f\x8b\b\x00\x00\x00\x00\x00\x00\xffRJ\xcb\xcfOJ,R\x02\x00\x00\x00\xff\xff\x00\x00\x00\xff\xff\x01\x00\x00\xff\xff\x81Z\x84\xc4\b\x00\x00\x00",
			errfn: errtest.NoError(),
		},
		{
			name: "gzip+base64",
			pf:   thrift.NewTBinaryProtocolFactoryDefault(),
			es: []encoding.Encoding{
				encoding.GZipEncoding,
				encoding.Base64Encoding,
			},
			in:    &stringTStruct{"foobar"},
			out:   "H4sIAAAAAAAA/2JgYGBLy89PSiwCAAAA//8AAAD//wEAAP//euNurwoA",
			errfn: errtest.NoError(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTSerializer(tt.pf, tt.es...)

			out, err := s.WriteString(tt.in)

			tt.errfn.Assert(t, err)
			assert.Equal(t, tt.out, out)
		})
	}
}
