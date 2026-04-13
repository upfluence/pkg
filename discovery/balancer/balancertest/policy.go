package balancertest

import (
	"context"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/discovery/resolver/static"
)

// PolicyFactory creates a new Policy instance for testing.
type PolicyFactory func() balancer.Policy[static.Peer]

// PolicyTest runs a comprehensive test suite for a Policy implementation.
// It tests common policy behaviors including:
// - Empty policy (no peers)
// - Single peer
// - Multiple peers
// - Adding and removing peers
func PolicyTest(t *testing.T, factory PolicyFactory) {
	for _, tt := range []struct {
		name string
		test func(*testing.T, PolicyFactory)
	}{
		{"NoPeers", testNoPeers},
		{"SinglePeer", testSinglePeer},
		{"AddAndRemovePeers", testAddAndRemovePeers},
		{"RemoveAllPeers", testRemoveAllPeers},
		{"AddPeersToEmpty", testAddPeersToEmpty},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, factory)
		})
	}
}

func testNoPeers(t *testing.T, factory PolicyFactory) {
	ctx := context.Background()
	policy := factory()

	// NoWait should return ErrNoPeerAvailable immediately
	p, done, err := policy.Get(ctx, balancer.GetOptions{NoWait: true})
	assert.Equal(t, balancer.ErrNoPeerAvailable, err)
	assert.Nil(t, done)
	assert.Empty(t, p.Addr())

	// Canceled context should return context.Canceled
	cctx, cancel := context.WithCancel(ctx)
	cancel()

	p, done, err = policy.Get(cctx, balancer.GetOptions{})
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, done)
	assert.Empty(t, p.Addr())
}

func testSinglePeer(t *testing.T, factory PolicyFactory) {
	ctx := context.Background()
	policy := factory()

	// Add a single peer
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{static.Peer("localhost:1")},
	})

	// Should get the same peer repeatedly
	for range 5 {
		p, done, err := policy.Get(ctx, balancer.GetOptions{})

		require.NoError(t, err)
		assert.NotNil(t, done)
		assert.Equal(t, "localhost:1", p.Addr())
		done(nil)
	}
}

func testAddAndRemovePeers(t *testing.T, factory PolicyFactory) {
	ctx := context.Background()
	policy := factory()

	// Add initial peers
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{
			static.Peer("localhost:1"),
			static.Peer("localhost:2"),
		},
	})

	// Verify we can get peers
	seen := make(map[string]bool)

	for range 50 {
		p, done, err := policy.Get(ctx, balancer.GetOptions{})

		require.NoError(t, err)
		assert.NotNil(t, done)

		seen[p.Addr()] = true

		done(nil)
	}

	assert.Contains(t, seen, "localhost:1")
	assert.Contains(t, seen, "localhost:2")

	// Update peers: remove localhost:1, add localhost:3
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{static.Peer("localhost:3")},
		Deletions: []static.Peer{static.Peer("localhost:1")},
	})

	// Verify we only see localhost:2 and localhost:3
	seen = make(map[string]bool)

	for range 50 {
		p, done, err := policy.Get(ctx, balancer.GetOptions{})

		require.NoError(t, err)
		assert.NotNil(t, done)

		seen[p.Addr()] = true

		done(nil)
	}

	assert.Contains(t, seen, "localhost:2")
	assert.Contains(t, seen, "localhost:3")
	assert.NotContains(t, seen, "localhost:1")
}

func testRemoveAllPeers(t *testing.T, factory PolicyFactory) {
	ctx := context.Background()
	policy := factory()

	// Add a peer
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{static.Peer("localhost:1")},
	})

	// Verify we can get it
	p, done, err := policy.Get(ctx, balancer.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "localhost:1", p.Addr())
	done(nil)

	// Remove the peer
	policy.Update(resolver.Update[static.Peer]{
		Deletions: []static.Peer{static.Peer("localhost:1")},
	})

	// NoWait should return ErrNoPeerAvailable
	p, done, err = policy.Get(ctx, balancer.GetOptions{NoWait: true})
	assert.Equal(t, balancer.ErrNoPeerAvailable, err)
	assert.Nil(t, done)
	assert.Empty(t, p.Addr())
}

func testAddPeersToEmpty(t *testing.T, factory PolicyFactory) {
	ctx := t.Context()
	policy := factory()

	// started is closed just before the goroutine enters Get's select,
	// giving us a deterministic signal instead of a sleep.
	started := make(chan struct{})
	done := make(chan struct{})

	go func() {
		close(started)

		p, doneFn, err := policy.Get(ctx, balancer.GetOptions{})
		assert.NoError(t, err)
		assert.NotNil(t, doneFn)
		assert.NotEmpty(t, p.Addr())
		doneFn(nil)
		close(done)
	}()

	// Wait until the goroutine has been scheduled, then yield once more so
	// it reaches the select inside Get before we call Update.
	<-started
	runtime.Gosched()

	// Add a peer — this closes the notifier and unblocks Get.
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{static.Peer("localhost:1")},
	})

	select {
	case <-done:
		// success
	case <-ctx.Done():
		t.Fatal("Get() did not unblock after adding peers")
	}
}
