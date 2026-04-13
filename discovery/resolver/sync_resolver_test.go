package resolver_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/discovery/resolver/static"
)

func TestNameResolverWithPeers(t *testing.T) {
	ctx := context.Background()
	nr := resolver.SyncResolverFromBuilder(
		static.Builder[static.Peer]{
			"n1": static.PeersFromStrings("foo", "bar"),
			"n2": static.PeersFromStrings("biz", "buz"),
		},
		false,
	)

	ps, err := nr.ResolveSync(ctx, "n1")

	require.NoError(t, err)
	assert.ElementsMatch(t, static.PeersFromStrings("foo", "bar"), ps)

	ps, err = nr.ResolveSync(ctx, "n2")

	require.NoError(t, err)
	assert.ElementsMatch(t, static.PeersFromStrings("biz", "buz"), ps)

	err = nr.Close()
	require.NoError(t, err)
}

func TestNameResolverNoPeerNoWait(t *testing.T) {
	ctx := context.Background()
	nr := resolver.SyncResolverFromBuilder(static.Builder[static.Peer]{}, true)

	ps, err := nr.ResolveSync(ctx, "n1")

	require.NoError(t, err)
	assert.ElementsMatch(t, 0, len(ps))

	err = nr.Close()
	require.NoError(t, err)
}

func TestNameResolverNoPeerWait(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	nr := resolver.SyncResolverFromBuilder(static.Builder[static.Peer]{}, false)

	defer cancel()

	ps, err := nr.ResolveSync(ctx, "n1")

	assert.Equal(t, context.DeadlineExceeded, err)
	assert.ElementsMatch(t, 0, len(ps))

	err = nr.Close()
	require.NoError(t, err)
}
