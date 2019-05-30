package serializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/encoding"
	"github.com/upfluence/pkg/testutil"
)

func TestContentType(t *testing.T) {
	for _, tt := range []struct {
		pf  thrift.TProtocolFactory
		es  []encoding.Encoding
		out string
	}{
		{
			pf:  thrift.NewTSimpleJSONProtocolFactory(),
			out: "application/json",
		},
		{
			pf:  thrift.NewTSimpleJSONProtocolFactory(),
			es:  []encoding.Encoding{encoding.Base64Encoding},
			out: "application/json+base64",
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
		errfn testutil.ErrorAssertion
	}{
		{
			name:  "regular encoding",
			pf:    thrift.NewTSimpleJSONProtocolFactory(),
			in:    &stringTStruct{"foobar"},
			out:   "\"foobar\"",
			errfn: testutil.NoError(),
		},
		{
			name:  "base64",
			pf:    thrift.NewTSimpleJSONProtocolFactory(),
			es:    []encoding.Encoding{encoding.Base64Encoding},
			in:    &stringTStruct{"foobar"},
			out:   "ImZvb2JhciI=",
			errfn: testutil.NoError(),
		},
		{
			name: "snappy",
			pf:   thrift.NewTSimpleJSONProtocolFactory(),
			es: []encoding.Encoding{
				encoding.SnappyEncoding,
			},
			in:    &stringTStruct{"foobar"},
			out:   "\xff\x06\x00\x00sNaPpY\x01\f\x00\x00\xff\x12\xfd\\\"foobar\"",
			errfn: testutil.NoError(),
		},
		{
			name: "snappy+base64",
			pf:   thrift.NewTBinaryProtocolFactoryDefault(),
			es: []encoding.Encoding{
				encoding.SnappyEncoding,
				encoding.Base64Encoding,
			},
			in:    &stringTStruct{"foobar"},
			out:   "/wYAAHNOYVBwWQEIAABlyOH6AAAABgEKAACWBYFbZm9vYmFy",
			errfn: testutil.NoError(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTSerializer(tt.pf, tt.es...)

			out, err := s.WriteString(tt.in)

			tt.errfn(t, err)
			assert.Equal(t, tt.out, out)
		})
	}
}
