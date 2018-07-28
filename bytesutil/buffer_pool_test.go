package bytesutil

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertEmptyBuffer(t *testing.T, b *bytes.Buffer) {
	assert.Equal(t, 0, b.Len(), "buffer has len() > 0")
	assert.Equal(t, "", b.String(), "buffer has content")
}

func TestIntegration(t *testing.T) {
	p := NewBufferPool()

	b := p.Get()
	assertEmptyBuffer(t, b)
	b.WriteString("foo")
	p.Put(b)

	b = p.Get()
	assertEmptyBuffer(t, b)
	p.Put(b)
}
