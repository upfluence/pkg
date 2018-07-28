package lru

import (
	"fmt"
	"sync"
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

func TestRaceCondition(t *testing.T) {
	c := NewCache(2)
	wg := &sync.WaitGroup{}

	for i := 0; i < 200; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for i := 0; i < 200; i++ {
				c.Add(fmt.Sprintf("buz %d", i), "bar")
				c.Get("foo")
			}
		}()
	}

	wg.Wait()
}
