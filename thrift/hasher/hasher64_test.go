package hasher

import (
	"hash"
	"hash/fnv"
	"io"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/v2/thrift/thrifttest"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type stringTStruct struct {
	s string
}

func (sts *stringTStruct) Write(p thrift.TProtocol) error {
	return p.WriteString(sts.s)
}

func (sts *stringTStruct) String() string {
	return sts.s
}

func (sts *stringTStruct) Read(p thrift.TProtocol) error {
	panic("unexpected")
}

type debugAccumulator struct {
	next accumulator

	log func(string, ...any)
}

func (da *debugAccumulator) beginOrdered() error {
	da.log("begin ordered context")
	return da.next.beginOrdered()
}

func (da *debugAccumulator) beginUnordered() error {
	da.log("begin unordered context")
	return da.next.beginUnordered()
}

func (da *debugAccumulator) endOrdered() error {
	da.log("end ordered context")
	return da.next.endOrdered()
}

func (da *debugAccumulator) endUnordered() error {
	da.log("end unordered context")
	return da.next.endUnordered()
}

func (da *debugAccumulator) Write(p []byte) (int, error) {
	if utf8.Valid(p) {
		da.log("write %d string: %q", len(p), string(p))
	} else {
		da.log("write %d bytes: %x", len(p), p)
	}

	return da.next.Write(p)
}

func (da *debugAccumulator) Close() error {
	da.log("close accumulator")

	return da.next.Close()
}

func TestHash(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   thrift.TStruct
		out  uint64
	}{
		{
			name: "string",
			in:   &stringTStruct{s: "foo"},
			out:  0xd8cbc7186ba13533,
		},
		{
			name: "list string",
			in: thrifttest.SliceWriter{
				&stringTStruct{s: "foo"},
				&stringTStruct{s: "bar"},
			},
			out: 0x3dc447b598597411,
		},
		{
			name: "map string string",
			in: thrifttest.MapWriter{
				"foo": &stringTStruct{s: "bar"},
				"bar": &stringTStruct{s: "foo"},
				"buz": &stringTStruct{s: "biz"},
			},
			out: 0xddac4f971fc7c18,
		},
		{
			name: "set string",
			in: thrifttest.SetWriter{
				&stringTStruct{s: "foo"},
				&stringTStruct{s: "bar"},
				&stringTStruct{s: "buz"},
			},
			out: 0xedca83a4cdd7df21,
		},
		{
			name: "set string",
			in: thrifttest.SetWriter{
				&stringTStruct{s: "buz"},
				&stringTStruct{s: "bar"},
				&stringTStruct{s: "foo"},
			},
			out: 0xedca83a4cdd7df21,
		},
		{
			name: "set string with missing element",
			in: thrifttest.SetWriter{
				&stringTStruct{s: "buz"},
				&stringTStruct{s: "bar"},
			},
			out: 0x7eada37398e53b95,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := newHasher64(
				fnv.New64,
				func(w io.Writer, hasherFunc func() hash.Hash) accumulator {
					return &debugAccumulator{
						next: newDeterministicAccumulator(w, hasherFunc),
						log: func(format string, args ...any) {
							t.Logf(format, args...)
						},
					}
				},
			)

			res, err := h.Hash(tt.in)

			assert.NoError(t, err)
			assert.Equal(t, tt.out, res)
		})
	}
}
