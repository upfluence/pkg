package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/v2/cache/policy/size"
)

func TestPolicyCache(t *testing.T) {
	c := WithEvictionPolicy[string](NewStringCache[string](), size.NewLRUPolicy[string](1))

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
