package simple

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"

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

func (p *roundRobinPicker) Pick(_ context.Context, peers []static.Peer) (static.Peer, error) {
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

	// Update is synchronous: peers are visible to Get immediately after return.
	policy.Update(resolver.Update[static.Peer]{Additions: peers})

	peer, _, err := policy.Get(context.Background(), balancer.GetOptions{NoWait: true})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

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
	// Update is synchronous: no sleep needed before Get.
	policy.Update(resolver.Update[static.Peer]{Additions: peers})

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

func (p *lastPicker) Pick(_ context.Context, peers []static.Peer) (static.Peer, error) {
	if len(peers) == 0 {
		return static.Peer(""), errors.New("no peers available")
	}

	return peers[len(peers)-1], nil
}

// errorPicker always returns an error
type errorPicker struct{}

func (p *errorPicker) Pick(_ context.Context, _ []static.Peer) (static.Peer, error) {
	return static.Peer(""), errors.New("picker error")
}

// TestRaceConditionPeerRemovalAfterWakeup verifies that Get() retries when
// peers are removed between the notifier being closed and Get() re-reading
// the peer list.
func TestRaceConditionPeerRemovalAfterWakeup(t *testing.T) {
	policy := NewPolicy(&roundRobinPicker{})
	ctx := t.Context()

	// started is closed just before the goroutine enters Get's select, giving
	// the test a deterministic signal to proceed rather than sleeping.
	started := make(chan struct{})
	gotPeer := make(chan static.Peer, 1)
	gotErr := make(chan error, 1)

	go func() {
		close(started)

		peer, _, err := policy.Get(ctx, balancer.GetOptions{})
		if err != nil {
			gotErr <- err
		} else {
			gotPeer <- peer
		}
	}()

	// Wait until the goroutine has been scheduled, then yield once more so it
	// reaches the select inside Get before we send any updates.
	<-started
	runtime.Gosched()

	// Add a peer — closes the notifier, waking the goroutine.
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{static.Peer("localhost:1")},
	})

	// Immediately remove it to exercise the retry path: the goroutine may have
	// woken from the notifier but not yet re-read the peer slice.
	policy.Update(resolver.Update[static.Peer]{
		Deletions: []static.Peer{static.Peer("localhost:1")},
	})

	// Re-add a peer so Get() can eventually succeed.
	// No sleep needed: Update() is synchronous and the goroutine will pick up
	// the new notifier on its next iteration.
	policy.Update(resolver.Update[static.Peer]{
		Additions: []static.Peer{static.Peer("localhost:2")},
	})

	select {
	case peer := <-gotPeer:
		addr := peer.Addr()
		if addr != "localhost:1" && addr != "localhost:2" {
			t.Errorf("Get() returned unexpected peer %q, want localhost:1 or localhost:2", addr)
		}
	case err := <-gotErr:
		t.Fatalf("Get() returned error %v, expected to retry and succeed", err)
	case <-ctx.Done():
		t.Fatal("Get() did not complete before test deadline")
	}
}
