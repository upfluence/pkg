package filter

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/discovery/resolver/resolvertest"
	"github.com/upfluence/pkg/v2/discovery/resolver/static"
)

func TestResolver(t *testing.T) {
	resolvertest.ResolverTest(t, func(peers []static.Peer) (resolver.Resolver[static.Peer], []static.Peer) {
		inner := static.NewResolver(peers)
		r := WrapResolver(inner, func(p static.Peer) bool {
			return strings.HasPrefix(p.Addr(), "localhost")
		})

		// Filter the expected peers
		var expected []static.Peer

		for _, p := range peers {
			if strings.HasPrefix(p.Addr(), "localhost") {
				expected = append(expected, p)
			}
		}

		return r, expected
	}, static.PeersFromStrings)
}

func TestFilterResolverAllowsAdditions(t *testing.T) {
	ctx := context.Background()

	inner := static.NewResolverFromStrings([]string{"allow:1", "deny:1"})
	r := WrapResolver(inner, func(p static.Peer) bool {
		return strings.HasPrefix(p.Addr(), "allow")
	})

	w := r.Resolve()

	u, err := w.Next(ctx, resolver.ResolveOptions{})

	require.NoError(t, err)
	assert.Equal(t, []static.Peer{static.Peer("allow:1")}, u.Additions)
	assert.Empty(t, u.Deletions)

	u, err = w.Next(ctx, resolver.ResolveOptions{NoWait: true})

	assert.Equal(t, resolver.ErrNoUpdates, err)
	assert.Equal(t, resolver.Update[static.Peer]{}, u)
}

func TestFilterResolverTracksDeletions(t *testing.T) {
	ctx := context.Background()

	inner := static.NewResolverFromStrings([]string{"allow:1", "deny:1"})
	r := WrapResolver(inner, func(p static.Peer) bool {
		return strings.HasPrefix(p.Addr(), "allow")
	})

	w := r.Resolve()

	_, err := w.Next(ctx, resolver.ResolveOptions{})
	require.NoError(t, err)

	inner.UpdatePeers(static.PeersFromStrings("allow:2", "deny:2"))

	u, err := w.Next(ctx, resolver.ResolveOptions{NoWait: true})

	require.NoError(t, err)
	assert.ElementsMatch(t, []static.Peer{static.Peer("allow:2")}, u.Additions)
	assert.ElementsMatch(t, []static.Peer{static.Peer("allow:1")}, u.Deletions)
}

func TestFilterResolverNoWaitFilteredEmpty(t *testing.T) {
	ctx := context.Background()

	inner := static.NewResolverFromStrings([]string{"deny:1"})
	r := WrapResolver(inner, func(p static.Peer) bool {
		return strings.HasPrefix(p.Addr(), "allow")
	})

	w := r.Resolve()

	u, err := w.Next(ctx, resolver.ResolveOptions{NoWait: true})

	assert.Equal(t, resolver.ErrNoUpdates, err)
	assert.Equal(t, resolver.Update[static.Peer]{}, u)

	inner.UpdatePeers(static.PeersFromStrings("allow:1"))

	u, err = w.Next(ctx, resolver.ResolveOptions{NoWait: true})

	require.NoError(t, err)
	assert.ElementsMatch(t, []static.Peer{static.Peer("allow:1")}, u.Additions)
	assert.Empty(t, u.Deletions)
}
