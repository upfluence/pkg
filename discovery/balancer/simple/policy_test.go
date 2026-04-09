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
