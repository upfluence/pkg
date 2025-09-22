package serializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/errors/errtest"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/v2/encoding"
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
			name: "gzip+base64",
			pf:   thrift.NewTBinaryProtocolFactoryDefault(),
			es: []encoding.Encoding{
				encoding.GZipEncoding,
				encoding.Base64Encoding,
			},
			in:    "H4sIAAAAAAAA/2JgYGBLy89PSiwCAAAA//8AAAD//wEAAP//euNurwoA",
			msgfn: func() TStruct { return &stringTStruct{} },
			out:   &stringTStruct{"foobar"},
			errfn: errtest.NoError(),
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
