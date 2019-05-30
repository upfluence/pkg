package serializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/encoding"
	"github.com/upfluence/pkg/testutil"
)

func TestDeserializerReadString(t *testing.T) {
	for _, tt := range []struct {
		name  string
		pf    thrift.TProtocolFactory
		es    []encoding.Encoding
		in    string
		msgfn func() TStruct
		out   TStruct
		errfn testutil.ErrorAssertion
	}{
		{
			name:  "regular encoding",
			pf:    thrift.NewTSimpleJSONProtocolFactory(),
			in:    "\"foobar\"",
			msgfn: func() TStruct { return &stringTStruct{} },
			out:   &stringTStruct{"foobar"},
			errfn: testutil.NoError(),
		},
		{
			name:  "base64",
			pf:    thrift.NewTSimpleJSONProtocolFactory(),
			es:    []encoding.Encoding{encoding.Base64Encoding},
			in:    "ImZvb2JhciI=",
			msgfn: func() TStruct { return &stringTStruct{} },
			out:   &stringTStruct{"foobar"},
			errfn: testutil.NoError(),
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
			errfn: testutil.NoError(),
		},
		{
			name: "snappy",
			pf:   thrift.NewTSimpleJSONProtocolFactory(),
			es: []encoding.Encoding{
				encoding.SnappyEncoding,
			},
			msgfn: func() TStruct { return &stringTStruct{} },
			out:   &stringTStruct{"foobar"},
			errfn: testutil.NoError(),
			in:    "\xff\x06\x00\x00sNaPpY\x01\f\x00\x00\xff\x12\xfd\\\"foobar\"",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := NewTDeserializer(tt.pf, tt.es...)

			msg := tt.msgfn()

			err := s.ReadString(msg, tt.in)

			tt.errfn(t, err)
			assert.Equal(t, tt.out, msg)
		})
	}
}
