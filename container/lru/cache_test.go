package lru

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertValue(t *testing.T, c *Cache, k Key, eV interface{}, eOk bool) {
	v, ok := c.Get("foo")

	assert.Equal(t, eV, v)
	assert.Equal(t, eOk, ok)
}

func TestIntegration(t *testing.T) {
	c := NewCache(2)

	assertValue(t, c, "foo", nil, false)
	c.Add("foo", "bar")
	assertValue(t, c, "foo", "bar", true)
	c.Add("fiz", "bar")
	assertValue(t, c, "foo", "bar", true)
	c.Add("buz", "bar")
	c.Add("bizz", "bar")
	assertValue(t, c, "foo", nil, false)
}
