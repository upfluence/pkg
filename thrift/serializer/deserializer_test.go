package serializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/errors/errtest"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/encoding"
)

func TestDeserializerReadString(t *testing.T) {
	for _, tt := range []struct {
		name  string
		pf    thrift.TProtocolFactory
		es    []encoding.Encoding
		in    string
		msgfn func() TStruct
		out   TStruct
		errfn errtest.ErrorAssertion
	}{
		{
			name:  "regular encoding",
			pf:    thrift.NewTJSONProtocolFactory(),
			in:    "\"foobar\"",
			msgfn: func() TStruct { return &stringTStruct{} },
			out:   &stringTStruct{"foobar"},
			errfn: errtest.NoError(),
		},
		{
			name:  "base64",
			pf:    thrift.NewTJSONProtocolFactory(),
			es:    []encoding.Encoding{encoding.Base64Encoding},
			in:    "ImZvb2JhciI=",
			msgfn: func() TStruct { return &stringTStruct{} },
			out:   &stringTStruct{"foobar"},
			errfn: errtest.NoError(),
		},
		{
			name: "snappy+base64",
			pf:   thrift.NewTBinaryProtocolFactoryDefault(),
			es: []encoding.Encoding{
				encoding.SnappyEncoding,
				encoding.Base64Encoding,
			},
			in:    "/wYAAHNOYVBwWQEIAABlyOH6AAAABgEKAACWBYFbZm9vYmFy",
			msgfn: func() TStruct { return &stringTStruct{} },
			out:   &stringTStruct{"foobar"},
			errfn: errtest.NoError(),
		},
		{
			name: "snappy",
			pf:   thrift.NewTJSONProtocolFactory(),
			es: []encoding.Encoding{
				encoding.SnappyEncoding,
			},
			msgfn: func() TStruct { return &stringTStruct{} },
			out:   &stringTStruct{"foobar"},
			errfn: errtest.NoError(),
			in:    "\xff\x06\x00\x00sNaPpY\x01\f\x00\x00\xff\x12\xfd\\\"foobar\"",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTDeserializer(tt.pf, tt.es...)

			msg := tt.msgfn()

			err := s.ReadString(msg, tt.in)

			tt.errfn.Assert(t, err)
			assert.Equal(t, tt.out, msg)
		})
	}
}
