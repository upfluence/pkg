package hasher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/v2/thrift/serializer"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type stringTStruct struct {
	s string
}

func (sts *stringTStruct) Write(p thrift.TProtocol) error {
	return p.WriteString(sts.s)
}

func (sts *stringTStruct) Read(p thrift.TProtocol) error {
	panic("unexpected")
}

func TestHash(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   serializer.TStruct
		out  uint64
	}{
		{
			name: "string",
			in:   &stringTStruct{s: "foo"},
			out:  0x81813814bc89acf6,
		},
		{
			name: "list string",
			in: serializer.SliceWriter{
				&stringTStruct{s: "foo"},
				&stringTStruct{s: "bar"},
			},
			out: 0xd4c58286fef8654,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFNVHasher64()

			res, err := h.Hash(tt.in)

			assert.NoError(t, err)
			assert.Equal(t, tt.out, res)
		})
	}
}
