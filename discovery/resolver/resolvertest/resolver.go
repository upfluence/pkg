package resolvertest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
)

// ResolverFactory creates a new Resolver instance for testing.
// The argument is the peers that should pre-exist.
// Returns the resolver and the peers that should actually be returned (after filtering/transformation).
type ResolverFactory[T peer.Peer] func([]T) (resolver.Resolver[T], []T)

// ResolverTest runs a comprehensive test suite for a Resolver implementation.
// It tests common resolver behaviors including:
// - Empty resolver (no initial peers)
// - Initial peers resolution
// - NoWait behavior with no updates
// - Context cancellation
// - Watcher closure
func ResolverTest[T peer.Peer](t *testing.T, factory ResolverFactory[T], makePeers func(...string) []T) {
	for _, tt := range []struct {
		name string
		test func(*testing.T, ResolverFactory[T], func(...string) []T)
	}{
		{"NoSeeds", testNoSeeds[T]},
		{"InitialPeers", testInitialPeers[T]},
		{"NoWaitNoUpdates", testNoWaitNoUpdates[T]},
		{"ContextCancellation", testContextCancellation[T]},
		{"WatcherClose", testWatcherClose[T]},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, factory, makePeers)
		})
	}
}

func testNoSeeds[T peer.Peer](t *testing.T, factory ResolverFactory[T], _ func(...string) []T) {
	ctx := context.Background()
	r, expected := factory(nil)

	assert.Empty(t, expected)

	require.NoError(t, r.Open(ctx))
	defer r.Close()

	w := r.Resolve()
	defer w.Close()

	// NoWait should return ErrNoUpdates immediately when no peers exist
	u, err := w.Next(ctx, resolver.ResolveOptions{NoWait: true})
	assert.Equal(t, resolver.ErrNoUpdates, err)
	assert.Empty(t, u.Additions)
	assert.Empty(t, u.Deletions)
}

func testInitialPeers[T peer.Peer](t *testing.T, factory ResolverFactory[T], makePeers func(...string) []T) {
	ctx := context.Background()
	peers := makePeers("localhost:1", "localhost:2")
	r, expected := factory(peers)

	require.NoError(t, r.Open(ctx))
	defer r.Close()

	w := r.Resolve()
	defer w.Close()

	// Should get initial peers
	u, err := w.Next(ctx, resolver.ResolveOptions{})
	require.NoError(t, err)
	assert.ElementsMatch(t, expected, u.Additions)
	assert.Empty(t, u.Deletions)
}

func testNoWaitNoUpdates[T peer.Peer](t *testing.T, factory ResolverFactory[T], makePeers func(...string) []T) {
	ctx := context.Background()
	peers := makePeers("localhost:1")
	r, _ := factory(peers)

	require.NoError(t, r.Open(ctx))
	defer r.Close()

	w := r.Resolve()
	defer w.Close()

	// Consume initial update
	_, err := w.Next(ctx, resolver.ResolveOptions{})
	require.NoError(t, err)

	// NoWait should return ErrNoUpdates when no updates are available
	u, err := w.Next(ctx, resolver.ResolveOptions{NoWait: true})
	assert.Equal(t, resolver.ErrNoUpdates, err)
	assert.Empty(t, u.Additions)
	assert.Empty(t, u.Deletions)
}

func testContextCancellation[T peer.Peer](t *testing.T, factory ResolverFactory[T], makePeers func(...string) []T) {
	ctx := context.Background()
	peers := makePeers("localhost:1")
	r, _ := factory(peers)

	require.NoError(t, r.Open(ctx))
	defer r.Close()

	w := r.Resolve()
	defer w.Close()

	// Consume initial update
	_, err := w.Next(ctx, resolver.ResolveOptions{})
	require.NoError(t, err)

	// Cancel context and try to wait for updates
	cctx, cancel := context.WithCancel(ctx)
	cancel()

	u, err := w.Next(cctx, resolver.ResolveOptions{})
	assert.Equal(t, context.Canceled, err)
	assert.Empty(t, u.Additions)
	assert.Empty(t, u.Deletions)
}

func testWatcherClose[T peer.Peer](t *testing.T, factory ResolverFactory[T], makePeers func(...string) []T) {
	ctx := context.Background()
	peers := makePeers("localhost:1")
	r, _ := factory(peers)

	require.NoError(t, r.Open(ctx))
	defer r.Close()

	w := r.Resolve()

	// Consume initial update
	_, err := w.Next(ctx, resolver.ResolveOptions{})
	require.NoError(t, err)

	// Close the watcher
	err = w.Close()
	require.NoError(t, err)

	// Trying to get next update should fail after close
	// Give a small timeout to avoid blocking forever
	cctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	_, err = w.Next(cctx, resolver.ResolveOptions{})
	assert.Error(t, err)
	// Should be either context.Canceled, context.DeadlineExceeded, or a close error
}
