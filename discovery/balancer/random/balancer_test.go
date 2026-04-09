package random

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/balancer/balancertest"
	"github.com/upfluence/pkg/v2/discovery/resolver/static"
)

func TestPolicy(t *testing.T) {
	balancertest.PolicyTest(t, func() balancer.Policy[static.Peer] {
		return NewPolicy[static.Peer]()
	})
}

func TestBalancerWithPeers(t *testing.T) {
	ctx := context.Background()
	b := NewBalancer(
		static.NewResolverFromStrings([]string{"localhost:0", "localhost:1", "localhost:2"}),
	)

	err := b.Open(ctx)
	assert.Nil(t, err)

	seen := make(map[string]int)
	for i := 0; i < 100; i++ {
		p, done, err := b.Get(ctx, balancer.GetOptions{})
		assert.Nil(t, err)
		assert.NotEmpty(t, p.Addr())
		seen[p.Addr()]++
		done(nil)
	}

	// With 100 requests across 3 peers, all should be selected at least once
	assert.Contains(t, seen, "localhost:0")
	assert.Contains(t, seen, "localhost:1")
	assert.Contains(t, seen, "localhost:2")

	b.Close()
}
