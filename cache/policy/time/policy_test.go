package time

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/v2/cache/policy"
)

func TestIdlePolicy(t *testing.T) {
	p := NewIdlePolicy[string](10 * time.Millisecond)

	p.Op("foo", policy.Set)
	p.Op("bar", policy.Set)
	p.Op("buz", policy.Set)
	p.Op("foo", policy.Get)
	p.Op("bar", policy.Evict)

	k := <-p.C()
	assert.Equal(t, "buz", k)
	k = <-p.C()
	assert.Equal(t, "foo", k)

	assert.Nil(t, p.Close())
}

func TestLifetimePolicy(t *testing.T) {
	p := NewLifetimePolicy[string](10 * time.Millisecond)

	p.Op("foo", policy.Set)
	p.Op("bar", policy.Set)
	p.Op("buz", policy.Set)
	p.Op("bar", policy.Get)
	p.Op("foo", policy.Evict)

	k := <-p.C()
	assert.Equal(t, "bar", k)
	k = <-p.C()
	assert.Equal(t, "buz", k)

	assert.Nil(t, p.Close())
}
