package transform_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/discovery/resolver/resolvertest"
	"github.com/upfluence/pkg/v2/discovery/resolver/static"
	"github.com/upfluence/pkg/v2/discovery/resolver/transform"
	"github.com/upfluence/pkg/v2/metadata"
)

type wrappedPeer struct {
	addr   string
	prefix string
}

func (p wrappedPeer) Addr() string                { return p.prefix + p.addr }
func (p wrappedPeer) Metadata() metadata.Metadata { return nil }

func makeWrappedPeers(addrs ...string) []wrappedPeer {
	peers := make([]wrappedPeer, len(addrs))
	for i, addr := range addrs {
		peers[i] = wrappedPeer{addr: addr, prefix: "prefix-"}
	}
	return peers
}

func TestResolver(t *testing.T) {
	resolvertest.ResolverTest(t, func(peers []wrappedPeer) (resolver.Resolver[wrappedPeer], []wrappedPeer) {
		// Extract source peers from wrapped peers
		sourcePeers := make([]static.Peer, len(peers))
		for i, p := range peers {
			// Remove the prefix to get back the original address
			sourcePeers[i] = static.Peer(p.addr)
		}

		source := static.NewResolver(sourcePeers)
		r := transform.WrapResolver(source, func(p static.Peer) wrappedPeer {
			return wrappedPeer{addr: p.Addr(), prefix: "prefix-"}
		})

		return r, peers
	}, makeWrappedPeers)
}

func TestTransformResolverTransformsUpdates(t *testing.T) {
	ctx := context.Background()
	source := static.NewResolverFromStrings([]string{"host1:80"})

	tr := transform.WrapResolver(source, func(p static.Peer) wrappedPeer {
		return wrappedPeer{addr: p.Addr(), prefix: "transformed-"}
	})

	assert.Nil(t, tr.Open(ctx))
	defer tr.Close()

	w := tr.Resolve()
	defer w.Close()

	// Get initial peers
	u, err := w.Next(ctx, resolver.ResolveOptions{})
	assert.Nil(t, err)
	assert.Len(t, u.Additions, 1)
	assert.Equal(t, "transformed-host1:80", u.Additions[0].Addr())

	// Update peers
	source.UpdatePeers(static.PeersFromStrings("host2:80", "host3:80"))

	u, err = w.Next(ctx, resolver.ResolveOptions{})
	assert.Nil(t, err)

	// Should have additions and deletions
	assert.Len(t, u.Additions, 2)
	assert.Len(t, u.Deletions, 1)

	additionAddrs := []string{u.Additions[0].Addr(), u.Additions[1].Addr()}
	assert.Contains(t, additionAddrs, "transformed-host2:80")
	assert.Contains(t, additionAddrs, "transformed-host3:80")

	assert.Equal(t, "transformed-host1:80", u.Deletions[0].Addr())
}

func TestTransformResolverMultipleWatchers(t *testing.T) {
	ctx := context.Background()
	source := static.NewResolverFromStrings([]string{"host1:80"})

	tr := transform.WrapResolver(source, func(p static.Peer) wrappedPeer {
		return wrappedPeer{addr: p.Addr(), prefix: "watcher-"}
	})

	assert.Nil(t, tr.Open(ctx))
	defer tr.Close()

	w1 := tr.Resolve()
	defer w1.Close()

	w2 := tr.Resolve()
	defer w2.Close()

	// Both watchers should get initial peers
	u1, err := w1.Next(ctx, resolver.ResolveOptions{})
	assert.Nil(t, err)
	assert.Len(t, u1.Additions, 1)
	assert.Equal(t, "watcher-host1:80", u1.Additions[0].Addr())

	u2, err := w2.Next(ctx, resolver.ResolveOptions{})
	assert.Nil(t, err)
	assert.Len(t, u2.Additions, 1)
	assert.Equal(t, "watcher-host1:80", u2.Additions[0].Addr())

	// Update peers
	source.UpdatePeers(static.PeersFromStrings("host2:80"))

	// Give time for update to propagate
	time.Sleep(10 * time.Millisecond)

	// Both watchers should receive the update
	u1, err = w1.Next(ctx, resolver.ResolveOptions{NoWait: true})
	assert.Nil(t, err)
	assert.Len(t, u1.Additions, 1)
	assert.Equal(t, "watcher-host2:80", u1.Additions[0].Addr())
	assert.Len(t, u1.Deletions, 1)
	assert.Equal(t, "watcher-host1:80", u1.Deletions[0].Addr())

	u2, err = w2.Next(ctx, resolver.ResolveOptions{NoWait: true})
	assert.Nil(t, err)
	assert.Len(t, u2.Additions, 1)
	assert.Equal(t, "watcher-host2:80", u2.Additions[0].Addr())
	assert.Len(t, u2.Deletions, 1)
	assert.Equal(t, "watcher-host1:80", u2.Deletions[0].Addr())
}
