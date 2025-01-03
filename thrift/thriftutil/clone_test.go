package thriftutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type stringTStruct struct {
	s string
}

func (sts *stringTStruct) String() string { return sts.s }

func (sts *stringTStruct) Write(p thrift.TProtocol) error {
	return p.WriteString(sts.s)
}

func (sts *stringTStruct) Read(p thrift.TProtocol) error {
	var err error

	sts.s, err = p.ReadString()

	return err
}

func TestClone(t *testing.T) {
	res, err := Clone(&stringTStruct{s: "foobar"})

	require.NoError(t, err)
	assert.Equal(t, &stringTStruct{"foobar"}, res)
}
