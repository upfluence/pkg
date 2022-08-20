package serializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/errors/errtest"
)

func TestMTDReadString(t *testing.T) {
	for _, tt := range []struct {
		ct, p string

		errfn errtest.ErrorAssertion
		out   string
	}{
		{
			p:     "\"foobar\"",
			errfn: errtest.NoError(),
			out:   "foobar",
		},
		{
			p:     "\"foobar\"",
			errfn: errtest.ErrorEqual(ErrProtocolNotProvided),
			ct:    "application/not-provided",
		},
		{
			p:     "\"foobar\"",
			errfn: errtest.ErrorEqual(ErrEncodingNotProvided),
			ct:    "application/json+not-provided",
		},
		{
			p:     "\"foobar\"",
			errfn: errtest.NoError(),
			out:   "foobar",
			ct:    "application/json",
		},
		{
			p:     "\x00\x00\x00\x06foobar",
			errfn: errtest.NoError(),
			out:   "foobar",
			ct:    "application/binary",
		},
		{
			p:     "\x1f\x8b\b\x00\x00\x00\x00\x00\x00\xffRJ\xcb\xcfOJ,R\x02\x00\x00\x00\xff\xff\x00\x00\x00\xff\xff\x01\x00\x00\xff\xff\x81Z\x84\xc4\b\x00\x00\x00",
			errfn: errtest.NoError(),
			out:   "foobar",
			ct:    "application/json+gzip",
		},
		{
			p:     "H4sIAAAAAAAA/2JgYGBLy89PSiwCAAAA//8AAAD//wEAAP//euNurwoA",
			errfn: errtest.NoError(),
			out:   "foobar",
			ct:    "application/binary+gzip+base64",
		},
	} {
		var (
			sts stringTStruct

			err = NewDefaultTMultiTypeDeserializer().ReadString(&sts, tt.p, tt.ct)
		)

		tt.errfn.Assert(t, err)
		assert.Equal(t, tt.out, sts.string)
	}
}
