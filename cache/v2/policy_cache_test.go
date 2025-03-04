package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/cache/v2/policy/size"
)

func TestPolicyCache(t *testing.T) {
	c := WithEvictionPolicy(NewCache[string, string](), size.NewLRUPolicy(1))

	done := make(chan struct{})

	testHookEviction = func() { close(done) }

	c.Set("foo", "bar")
	c.Set("bar", "buz")

	<-done

	_, ok, err := c.Get("foo")

	assert.False(t, ok)
	assert.Nil(t, err)

	assert.Nil(t, c.Close())

	testHookEviction = nop
}
