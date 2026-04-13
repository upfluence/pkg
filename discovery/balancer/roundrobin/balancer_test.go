package roundrobin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/upfluence/pkg/v2/discovery/balancer"
	"github.com/upfluence/pkg/v2/discovery/balancer/balancertest"
	"github.com/upfluence/pkg/v2/discovery/resolver/static"
)

func TestPolicy(t *testing.T) {
	balancertest.PolicyTest(t, func() balancer.Policy[static.Peer] {
		return NewPolicy[static.Peer]()
	})
}

func TestBalanceRoundRobinOrder(t *testing.T) {
	ctx := context.Background()
	b := NewBalancer(
		static.NewResolverFromStrings([]string{"localhost:0", "localhost:1", "localhost:2"}),
	)

	err := b.Open(ctx)
	require.NoError(t, err)

	// Collect the first full cycle. The initial Get blocks until the Puller
	// goroutine delivers peers, so this also acts as the synchronisation point.
	// The policy preserves insertion order, so subsequent cycles are identical.
	var firstCycle [3]string

	for i := range 3 {
		p, done, err := b.Get(ctx, balancer.GetOptions{})

		require.NoError(t, err)

		firstCycle[i] = p.Addr()

		done(nil)
	}

	// All three peers must appear in the first cycle.
	assert.ElementsMatch(t, []string{"localhost:0", "localhost:1", "localhost:2"}, firstCycle[:])

	// Subsequent cycles must repeat in the exact same order.
	for range 2 {
		for i := range 3 {
			p, done, err := b.Get(ctx, balancer.GetOptions{})

			require.NoError(t, err)
			assert.Equal(t, firstCycle[i], p.Addr())

			done(nil)
		}
	}

	b.Close()
}
