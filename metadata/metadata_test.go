package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringPassing(t *testing.T) {
	md := Metadata{}

	err := Encode(md, "Foo", "bar")
	assert.Nil(t, err)

	var got string
	Decode(md, "foo", &got)

	assert.Equal(t, "bar", got)
}

func TestPais(t *testing.T) {
	md := Pairs("foo", "bar", "buz", "biz")

	assert.Equal(t, 2, len(md))
	assert.Equal(t, "bar", md.Fetch("foo"))
	assert.Equal(t, "biz", md.Fetch("buz"))
}
