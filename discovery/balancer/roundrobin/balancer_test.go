package roundrobin

import (
	"context"
	"testing"
	"time"

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

func TestBalanceRoundRobinOrder(t *testing.T) {
	ctx := context.Background()
	b := NewBalancer(
		static.NewResolverFromStrings([]string{"localhost:0", "localhost:1", "localhost:2"}),
	)

	err := b.Open(ctx)
	assert.Nil(t, err)

	time.Sleep(10 * time.Millisecond)

	// Verify round-robin cycles through all peers in order
	for cycle := 0; cycle < 3; cycle++ {
		for i := 0; i < 3; i++ {
			p, done, err := b.Get(ctx, balancer.GetOptions{})
			assert.Nil(t, err)
			assert.Contains(t, []string{"localhost:0", "localhost:1", "localhost:2"}, p.Addr())
			done(nil)
		}
	}

	b.Close()
}
