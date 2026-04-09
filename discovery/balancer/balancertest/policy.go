package balancertest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
	for i := 0; i < 5; i++ {
		p, done, err := policy.Get(ctx, balancer.GetOptions{})
		assert.Nil(t, err)
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
	for i := 0; i < 50; i++ {
		p, done, err := policy.Get(ctx, balancer.GetOptions{})
		assert.Nil(t, err)
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
	for i := 0; i < 50; i++ {
		p, done, err := policy.Get(ctx, balancer.GetOptions{})
		assert.Nil(t, err)
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
	assert.Nil(t, err)
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
	ctx := context.Background()
	policy := factory()

	// Start with no peers, spawn a goroutine that will wait
	done := make(chan struct{})
	go func() {
		p, doneFn, err := policy.Get(ctx, balancer.GetOptions{})
		assert.Nil(t, err)
		assert.NotNil(t, doneFn)
		assert.NotEmpty(t, p.Addr())
		doneFn(nil)
		close(done)
	}()

	// Give the goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Add a peer
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{static.Peer("localhost:1")},
	})

	// The waiting goroutine should complete
	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Get() did not unblock after adding peers")
	}
}
