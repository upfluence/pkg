package balancer_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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

type testPolicy struct {
	mu      sync.Mutex
	peers   []wrappedPeer
	updates []resolver.Update[wrappedPeer]
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
	defer p.mu.Unlock()

	p.updates = append(p.updates, u)

	for _, peer := range u.Additions {
		p.peers = append(p.peers, peer)
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
	ctx := context.Background()
	r := static.NewResolverFromStrings([]string{"localhost:1", "localhost:2"})
	policy := &testPolicy{}

	b := balancer.WrapPolicy(
		r,
		policy,
		func(sp static.Peer) (wrappedPeer, error) {
			return wrappedPeer{addr: sp.Addr()}, nil
		},
	)

	assert.Nil(t, b.Open(ctx))

	time.Sleep(10 * time.Millisecond)

	updates := policy.getUpdates()
	assert.Len(t, updates, 1)
	assert.ElementsMatch(t, []wrappedPeer{
		{addr: "localhost:1"},
		{addr: "localhost:2"},
	}, updates[0].Additions)
	assert.Empty(t, updates[0].Deletions)

	peer, done, err := b.Get(ctx, balancer.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, "localhost:1", peer.Addr())
	done(nil)

	assert.Nil(t, b.Close())
}

func TestWrapPolicyMapsDeletions(t *testing.T) {
	ctx := context.Background()
	r := static.NewResolverFromStrings([]string{"localhost:1", "localhost:2"})
	policy := &testPolicy{}

	b := balancer.WrapPolicy(
		r,
		policy,
		func(sp static.Peer) (wrappedPeer, error) {
			return wrappedPeer{addr: sp.Addr()}, nil
		},
	)

	assert.Nil(t, b.Open(ctx))
	time.Sleep(10 * time.Millisecond)

	r.UpdatePeers(static.PeersFromStrings("localhost:2", "localhost:3"))
	time.Sleep(10 * time.Millisecond)

	updates := policy.getUpdates()
	assert.Len(t, updates, 2)

	assert.ElementsMatch(t, []wrappedPeer{{addr: "localhost:3"}}, updates[1].Additions)
	assert.ElementsMatch(t, []wrappedPeer{{addr: "localhost:1"}}, updates[1].Deletions)

	assert.Nil(t, b.Close())
}

func TestWrapPolicySkipsFailedBuilds(t *testing.T) {
	ctx := context.Background()
	r := static.NewResolverFromStrings([]string{"localhost:1", "fail:2", "localhost:3"})
	policy := &testPolicy{}

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

	assert.Nil(t, b.Open(ctx))
	time.Sleep(10 * time.Millisecond)

	updates := policy.getUpdates()
	assert.Len(t, updates, 1)
	assert.ElementsMatch(t, []wrappedPeer{
		{addr: "localhost:1"},
		{addr: "localhost:3"},
	}, updates[0].Additions)

	assert.Nil(t, b.Close())
}

func TestWrapPolicyDelegatesGetToPolicy(t *testing.T) {
	ctx := context.Background()
	r := static.NewResolverFromStrings([]string{"localhost:1"})
	policy := &testPolicy{}

	b := balancer.WrapPolicy(
		r,
		policy,
		func(sp static.Peer) (wrappedPeer, error) {
			return wrappedPeer{addr: sp.Addr()}, nil
		},
	)

	assert.Nil(t, b.Open(ctx))
	time.Sleep(10 * time.Millisecond)

	peer, done, err := b.Get(ctx, balancer.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, "localhost:1", peer.Addr())
	done(nil)

	peer, done, err = b.Get(ctx, balancer.GetOptions{NoWait: true})
	assert.Equal(t, balancer.ErrNoPeerAvailable, err)
	assert.Nil(t, done)

	assert.Nil(t, b.Close())
}
