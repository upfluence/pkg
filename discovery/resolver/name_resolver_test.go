package resolver_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/discovery/resolver"
	"github.com/upfluence/pkg/discovery/resolver/static"
)

func TestNameResolver(t *testing.T) {
	ctx := context.Background()
	nr := resolver.NameResolver{
		Builder: static.Builder{
			"n1": static.PeersFromStrings("foo", "bar"),
			"n2": static.PeersFromStrings("biz", "buz"),
		},
	}

	ps, err := nr.Resolve(ctx, "n1")

	assert.Nil(t, err)
	assert.ElementsMatch(t, static.PeersFromStrings("foo", "bar"), ps)

	ps, err = nr.Resolve(ctx, "n2")

	assert.Nil(t, err)
	assert.ElementsMatch(t, static.PeersFromStrings("biz", "buz"), ps)

	err = nr.Close()
	assert.Nil(t, err)
}
