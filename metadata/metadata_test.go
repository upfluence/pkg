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
