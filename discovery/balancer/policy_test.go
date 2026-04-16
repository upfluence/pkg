package balancer_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/discovery/resolver/static"
	"github.com/upfluence/pkg/v2/metadata"
)

type wrappedPeer struct {
	addr string
}

func (p wrappedPeer) Addr() string                { return p.addr }
func (p wrappedPeer) Metadata() metadata.Metadata { return nil }

// testPolicy is a Policy implementation used in tests that records all
// updates it receives and exposes waitForUpdate to synchronise on them
// without relying on time.Sleep.
type testPolicy struct {
	mu      sync.Mutex
	peers   []wrappedPeer
	updates []resolver.Update[wrappedPeer]
	// updatec is closed on each Update call and immediately replaced so
	// that waitForUpdate returns exactly once per Update.
	updatec chan struct{}
}

func newTestPolicy() *testPolicy {
	return &testPolicy{updatec: make(chan struct{})}
}

func (p *testPolicy) Get(ctx context.Context, opts balancer.GetOptions) (wrappedPeer, func(error), error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.peers) == 0 {
		if opts.NoWait {
			return wrappedPeer{}, nil, balancer.ErrNoPeerAvailable
		}

		<-ctx.Done()

		return wrappedPeer{}, nil, ctx.Err()
	}

	peer := p.peers[0]
	p.peers = p.peers[1:]

	return peer, func(error) {}, nil
}

func (p *testPolicy) Update(u resolver.Update[wrappedPeer]) {
	p.mu.Lock()

	p.updates = append(p.updates, u)
	p.peers = append(p.peers, u.Additions...)

	// Close the current channel and replace it before releasing the lock so
	// waitForUpdate never misses a notification.
	ch := p.updatec
	p.updatec = make(chan struct{})

	p.mu.Unlock()

	close(ch)
}

// waitForUpdate blocks until the next Update call completes or ctx is done.
func (p *testPolicy) waitForUpdate(ctx context.Context) error {
	p.mu.Lock()
	ch := p.updatec
	p.mu.Unlock()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *testPolicy) getUpdates() []resolver.Update[wrappedPeer] {
	p.mu.Lock()
	defer p.mu.Unlock()

	updates := make([]resolver.Update[wrappedPeer], len(p.updates))
	copy(updates, p.updates)

	return updates
}

func TestWrapPolicyMapsAdditions(t *testing.T) {
	ctx := t.Context()
	r := static.NewResolverFromStrings([]string{"localhost:1", "localhost:2"})
	policy := newTestPolicy()

	b := balancer.WrapPolicy(
		r,
		policy,
		func(sp static.Peer) (wrappedPeer, error) {
			return wrappedPeer{addr: sp.Addr()}, nil
		},
	)

	require.NoError(t, b.Open(ctx))
	require.NoError(t, policy.waitForUpdate(ctx))

	updates := policy.getUpdates()
	assert.Len(t, updates, 1)
	assert.ElementsMatch(t, []wrappedPeer{
		{addr: "localhost:1"},
		{addr: "localhost:2"},
	}, updates[0].Additions)
	assert.Empty(t, updates[0].Deletions)

	peer, done, err := b.Get(ctx, balancer.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "localhost:1", peer.Addr())
	done(nil)

	assert.NoError(t, b.Close())
}

func TestWrapPolicyMapsDeletions(t *testing.T) {
	ctx := t.Context()
	r := static.NewResolverFromStrings([]string{"localhost:1", "localhost:2"})
	policy := newTestPolicy()

	b := balancer.WrapPolicy(
		r,
		policy,
		func(sp static.Peer) (wrappedPeer, error) {
			return wrappedPeer{addr: sp.Addr()}, nil
		},
	)

	require.NoError(t, b.Open(ctx))
	require.NoError(t, policy.waitForUpdate(ctx)) // initial peers

	r.UpdatePeers(static.PeersFromStrings("localhost:2", "localhost:3"))
	require.NoError(t, policy.waitForUpdate(ctx)) // diff update

	updates := policy.getUpdates()
	assert.Len(t, updates, 2)

	assert.ElementsMatch(t, []wrappedPeer{{addr: "localhost:3"}}, updates[1].Additions)
	assert.ElementsMatch(t, []wrappedPeer{{addr: "localhost:1"}}, updates[1].Deletions)

	assert.NoError(t, b.Close())
}

func TestWrapPolicySkipsFailedBuilds(t *testing.T) {
	ctx := t.Context()
	r := static.NewResolverFromStrings([]string{"localhost:1", "fail:2", "localhost:3"})
	policy := newTestPolicy()

	b := balancer.WrapPolicy(
		r,
		policy,
		func(sp static.Peer) (wrappedPeer, error) {
			if sp.Addr() == "fail:2" {
				return wrappedPeer{}, errors.New("build failed")
			}

			return wrappedPeer{addr: sp.Addr()}, nil
		},
	)

	require.NoError(t, b.Open(ctx))
	require.NoError(t, policy.waitForUpdate(ctx))

	updates := policy.getUpdates()
	assert.Len(t, updates, 1)
	assert.ElementsMatch(t, []wrappedPeer{
		{addr: "localhost:1"},
		{addr: "localhost:3"},
	}, updates[0].Additions)

	assert.NoError(t, b.Close())
}

func TestWrapPolicyDelegatesGetToPolicy(t *testing.T) {
	ctx := t.Context()
	r := static.NewResolverFromStrings([]string{"localhost:1"})
	policy := newTestPolicy()

	b := balancer.WrapPolicy(
		r,
		policy,
		func(sp static.Peer) (wrappedPeer, error) {
			return wrappedPeer{addr: sp.Addr()}, nil
		},
	)

	require.NoError(t, b.Open(ctx))
	require.NoError(t, policy.waitForUpdate(ctx))

	peer, done, err := b.Get(ctx, balancer.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "localhost:1", peer.Addr())
	done(nil)

	_, _, err = b.Get(ctx, balancer.GetOptions{NoWait: true})
	assert.Equal(t, balancer.ErrNoPeerAvailable, err)

	assert.NoError(t, b.Close())
}
