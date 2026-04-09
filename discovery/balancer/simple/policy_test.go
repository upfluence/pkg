package simple

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/balancer/balancertest"
	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/discovery/resolver/static"
)

func TestPolicy(t *testing.T) {
	balancertest.PolicyTest(t, func() balancer.Policy[static.Peer] {
		return NewPolicy(&roundRobinPicker{})
	})
}

// roundRobinPicker cycles through peers sequentially
type roundRobinPicker struct {
	mu    sync.Mutex
	index int
}

func (p *roundRobinPicker) Pick(ctx context.Context, peers []static.Peer) (static.Peer, error) {
	if len(peers) == 0 {
		return static.Peer(""), errors.New("no peers available")
	}

	p.mu.Lock()
	idx := p.index % len(peers)
	p.index++
	p.mu.Unlock()

	return peers[idx], nil
}

func TestPickerDelegation(t *testing.T) {
	policy := NewPolicy(&lastPicker{})

	peers := []static.Peer{
		static.Peer("peer1"),
		static.Peer("peer2"),
		static.Peer("peer3"),
	}

	policy.Update(resolver.Update[static.Peer]{Additions: peers})
	time.Sleep(10 * time.Millisecond)

	peer, _, err := policy.Get(context.Background(), balancer.GetOptions{NoWait: true})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Verify we got one of the peers (order from map is not guaranteed)
	found := false
	for _, p := range peers {
		if peer.Addr() == p.Addr() {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Get() returned unexpected peer %q", peer.Addr())
	}
}

func TestPickerError(t *testing.T) {
	policy := NewPolicy(&errorPicker{})

	peers := []static.Peer{static.Peer("peer1")}
	policy.Update(resolver.Update[static.Peer]{Additions: peers})
	time.Sleep(10 * time.Millisecond)

	_, _, err := policy.Get(context.Background(), balancer.GetOptions{NoWait: true})
	if err == nil {
		t.Fatal("Get() expected error, got nil")
	}

	if err.Error() != "picker error" {
		t.Errorf("Get() error = %q, want %q", err.Error(), "picker error")
	}
}

// lastPicker picks the last peer from the list
type lastPicker struct{}

func (p *lastPicker) Pick(ctx context.Context, peers []static.Peer) (static.Peer, error) {
	if len(peers) == 0 {
		return static.Peer(""), errors.New("no peers available")
	}
	return peers[len(peers)-1], nil
}

// errorPicker always returns an error
type errorPicker struct{}

func (p *errorPicker) Pick(ctx context.Context, peers []static.Peer) (static.Peer, error) {
	return static.Peer(""), errors.New("picker error")
}

// TestRaceConditionPeerRemovalAfterWakeup tests the race condition where
// peers are removed after Get() wakes up from the notifier but before it
// re-reads the peer list. The fix should retry waiting in this case.
func TestRaceConditionPeerRemovalAfterWakeup(t *testing.T) {
	policy := NewPolicy(&roundRobinPicker{})
	ctx := context.Background()

	// Start a goroutine that will wait for peers
	gotPeer := make(chan static.Peer)
	gotErr := make(chan error)
	go func() {
		peer, _, err := policy.Get(ctx, balancer.GetOptions{})
		if err != nil {
			gotErr <- err
		} else {
			gotPeer <- peer
		}
	}()

	// Give the goroutine time to start waiting
	time.Sleep(10 * time.Millisecond)

	// Add a peer (this will close the notifier)
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{static.Peer("localhost:1")},
	})

	// Immediately remove the peer to simulate the race condition
	// This happens after the notifier is closed but potentially before
	// Get() re-reads the peer list
	policy.Update(resolver.Update[static.Peer]{
		Deletions: []static.Peer{static.Peer("localhost:1")},
	})

	// Add the peer back so Get() can eventually succeed
	time.Sleep(5 * time.Millisecond)
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{static.Peer("localhost:2")},
	})

	// The Get() call should eventually succeed (not return error)
	// It could get either localhost:1 (if fast enough) or localhost:2 (after retry)
	select {
	case peer := <-gotPeer:
		addr := peer.Addr()
		if addr != "localhost:1" && addr != "localhost:2" {
			t.Errorf("Get() returned unexpected peer %q, want localhost:1 or localhost:2", addr)
		}
	case err := <-gotErr:
		t.Fatalf("Get() returned error %v, expected to retry and succeed", err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Get() did not complete in time")
	}
}
