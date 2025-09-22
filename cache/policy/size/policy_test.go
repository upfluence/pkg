package size

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/cache/policy"
)

func TestLRUPolicy(t *testing.T) {
	p := NewLRUPolicy[string](2)

	p.Op("foo", policy.Set)
	p.Op("bar", policy.Set)
	p.Op("buz", policy.Set)

	k := <-p.C()
	assert.Equal(t, "foo", k)

	p.Op("bar", policy.Get)
	p.Op("foo", policy.Set)

	k = <-p.C()
	assert.Equal(t, "buz", k)

	p.Close()

	assert.Equal(t, policy.ErrClosed, p.Op("foo", policy.Set))
}
