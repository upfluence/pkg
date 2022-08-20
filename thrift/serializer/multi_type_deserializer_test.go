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
			p:     "\xff\x06\x00\x00sNaPpY\x01\f\x00\x00\xff\x12\xfd\\\"foobar\"",
			errfn: errtest.NoError(),
			out:   "foobar",
			ct:    "application/json+snappy",
		},
		{
			p:     "/wYAAHNOYVBwWQEIAABlyOH6AAAABgEKAACWBYFbZm9vYmFy",
			errfn: errtest.NoError(),
			out:   "foobar",
			ct:    "application/binary+snappy+base64",
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
