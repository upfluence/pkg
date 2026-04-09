package static

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/discovery/resolver/resolvertest"
)

func TestResolver(t *testing.T) {
	resolvertest.ResolverTest(t, func(peers []Peer) (resolver.Resolver[Peer], []Peer) {
		return NewResolver(peers), peers
	}, PeersFromStrings)
}

func TestResolve(t *testing.T) {
	ctx := context.Background()
	r := NewResolverFromStrings([]string{"localhost:1", "localhost:2"})

	w := r.Resolve()

	u, err := w.Next(ctx, resolver.ResolveOptions{})

	assert.Nil(t, err)
	assert.Equal(
		t,
		resolver.Update[Peer]{
			Additions: []Peer{
				Peer("localhost:1"),
				Peer("localhost:2"),
			},
		},
		u,
	)

	u, err = w.Next(ctx, resolver.ResolveOptions{NoWait: true})

	assert.Equal(t, err, resolver.ErrNoUpdates)
	assert.Equal(t, resolver.Update[Peer]{}, u)

	err = w.Close()
	assert.Nil(t, err)

	u, err = w.Next(ctx, resolver.ResolveOptions{})

	assert.Equal(t, err, context.Canceled)
	assert.Equal(t, resolver.Update[Peer]{}, u)
}

func TestPeers(t *testing.T) {
	r := NewResolverFromStrings([]string{"localhost:1", "localhost:2"})

	peers := r.Peers()

	assert.Equal(t, []Peer{Peer("localhost:1"), Peer("localhost:2")}, peers)

	// Mutating the returned slice should not affect the resolver
	peers[0] = Peer("localhost:99")
	assert.Equal(t, []Peer{Peer("localhost:1"), Peer("localhost:2")}, r.Peers())
}

func TestPeersEmpty(t *testing.T) {
	r := NewResolver[Peer](nil)
	assert.Equal(t, []Peer{}, r.Peers())
}

func TestUpdatePeers(t *testing.T) {
	r := NewResolverFromStrings([]string{"localhost:1", "localhost:2"})

	w := r.Resolve()

	// Consume initial update
	u, err := w.Next(context.Background(), resolver.ResolveOptions{})
	assert.Nil(t, err)
	assert.Equal(
		t,
		resolver.Update[Peer]{
			Additions: []Peer{
				Peer("localhost:1"),
				Peer("localhost:2"),
			},
		},
		u,
	)

	// Update peers: remove localhost:1, add localhost:3
	r.UpdatePeers(PeersFromStrings("localhost:2", "localhost:3"))

	assert.Equal(
		t,
		[]Peer{Peer("localhost:2"), Peer("localhost:3")},
		r.Peers(),
	)

	// The watcher should receive the diff
	u, err = w.Next(context.Background(), resolver.ResolveOptions{NoWait: true})
	assert.Nil(t, err)
	assert.ElementsMatch(t, []Peer{Peer("localhost:3")}, u.Additions)
	assert.ElementsMatch(t, []Peer{Peer("localhost:1")}, u.Deletions)
}

func TestUpdatePeersNoChange(t *testing.T) {
	r := NewResolverFromStrings([]string{"localhost:1", "localhost:2"})

	w := r.Resolve()

	// Consume initial update
	_, err := w.Next(context.Background(), resolver.ResolveOptions{})
	assert.Nil(t, err)

	// Update with same peers — no diff
	r.UpdatePeers(PeersFromStrings("localhost:1", "localhost:2"))

	// Should have no update
	_, err = w.Next(context.Background(), resolver.ResolveOptions{NoWait: true})
	assert.Equal(t, resolver.ErrNoUpdates, err)
}

func TestUpdatePeersMultipleWatchers(t *testing.T) {
	r := NewResolverFromStrings([]string{"localhost:1"})

	w1 := r.Resolve()
	w2 := r.Resolve()

	// Consume initial updates
	_, err := w1.Next(context.Background(), resolver.ResolveOptions{})
	assert.Nil(t, err)
	_, err = w2.Next(context.Background(), resolver.ResolveOptions{})
	assert.Nil(t, err)

	r.UpdatePeers(PeersFromStrings("localhost:1", "localhost:2"))

	u1, err := w1.Next(context.Background(), resolver.ResolveOptions{NoWait: true})
	assert.Nil(t, err)
	assert.ElementsMatch(t, []Peer{Peer("localhost:2")}, u1.Additions)

	u2, err := w2.Next(context.Background(), resolver.ResolveOptions{NoWait: true})
	assert.Nil(t, err)
	assert.ElementsMatch(t, []Peer{Peer("localhost:2")}, u2.Additions)
}

func TestUpdatePeersClosedWatcher(t *testing.T) {
	r := NewResolverFromStrings([]string{"localhost:1"})

	w := r.Resolve()

	// Consume initial update
	_, err := w.Next(context.Background(), resolver.ResolveOptions{})
	assert.Nil(t, err)

	// Close the watcher — it should unsubscribe
	err = w.Close()
	assert.Nil(t, err)

	// This should not block or panic
	r.UpdatePeers(PeersFromStrings("localhost:2"))
}

func TestUpdatePeersBlockingWatcher(t *testing.T) {
	r := NewResolverFromStrings([]string{"localhost:1"})

	w := r.Resolve()

	// Consume initial update
	_, err := w.Next(context.Background(), resolver.ResolveOptions{})
	assert.Nil(t, err)

	// Start a blocking Next call in a goroutine
	done := make(chan resolver.Update[Peer], 1)
	go func() {
		u, err := w.Next(context.Background(), resolver.ResolveOptions{})
		assert.Nil(t, err)
		done <- u
	}()

	// Update peers — should unblock the watcher
	r.UpdatePeers(PeersFromStrings("localhost:2"))

	u := <-done
	assert.ElementsMatch(t, []Peer{Peer("localhost:2")}, u.Additions)
	assert.ElementsMatch(t, []Peer{Peer("localhost:1")}, u.Deletions)
}
