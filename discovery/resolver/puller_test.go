package resolver_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/discovery/resolver"
	"github.com/upfluence/pkg/v2/discovery/resolver/static"
)

type transientErrorStaticResolver struct {
	inner    *static.Resolver[static.Peer]
	failed   atomic.Bool
	failWith error
}

func newTransientErrorStaticResolver(peers []string, err error) *transientErrorStaticResolver {
	return &transientErrorStaticResolver{
		inner:    static.NewResolverFromStrings(peers),
		failWith: err,
	}
}

func (r *transientErrorStaticResolver) Open(ctx context.Context) error {
	return r.inner.Open(ctx)
}

func (r *transientErrorStaticResolver) Close() error {
	return r.inner.Close()
}

func (r *transientErrorStaticResolver) Resolve() resolver.Watcher[static.Peer] {
	return &transientErrorWatcher{
		inner:    r.inner.Resolve(),
		failed:   &r.failed,
		failWith: r.failWith,
	}
}

func (r *transientErrorStaticResolver) UpdatePeers(peers []static.Peer) {
	r.inner.UpdatePeers(peers)
}

type transientErrorWatcher struct {
	inner    resolver.Watcher[static.Peer]
	failed   *atomic.Bool
	failWith error
}

func (w *transientErrorWatcher) Next(ctx context.Context, opts resolver.ResolveOptions) (resolver.Update[static.Peer], error) {
	if w.failed.CompareAndSwap(false, true) {
		return resolver.Update[static.Peer]{}, w.failWith
	}

	return w.inner.Next(ctx, opts)
}

func (w *transientErrorWatcher) Close() error {
	return w.inner.Close()
}

func TestPullerRecoversAfterWatcherError(t *testing.T) {
	r := newTransientErrorStaticResolver([]string{"allow:1"}, errors.New("boom"))

	updates := make(chan resolver.Update[static.Peer], 1)
	p, _ := resolver.NewPuller(r, func(u resolver.Update[static.Peer]) {
		if len(u.Additions) > 0 || len(u.Deletions) > 0 {
			updates <- u
		}
	})

	require.NoError(t, p.Open(context.Background()))

	select {
	case u := <-updates:
		assert.Equal(t, []static.Peer{static.Peer("allow:1")}, u.Additions)
	case <-t.Context().Done():
		t.Fatal("timed out waiting for update after watcher error")
	}

	assert.NoError(t, p.Close())
}

func TestPullerIsOpenTracksClose(t *testing.T) {
	r := static.NewResolverFromStrings([]string{"allow:1"})
	p, _ := resolver.NewPuller(r, func(resolver.Update[static.Peer]) {})
	assert.False(t, p.IsOpen())

	require.NoError(t, p.Open(context.Background()))
	assert.True(t, p.IsOpen())

	require.NoError(t, p.Close())
	assert.False(t, p.IsOpen())
}
