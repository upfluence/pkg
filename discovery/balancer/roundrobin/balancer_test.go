package roundrobin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/pkg/discovery/balancer"
	"github.com/upfluence/pkg/discovery/resolver/static"
)

func TestBalanceEmpty(t *testing.T) {
	ctx := context.Background()
	b := NewBalancer(&static.Resolver{})

	p, err := b.Get(ctx, balancer.GetOptions{NoWait: true})

	assert.Nil(t, p)
	assert.Equal(t, balancer.ErrNoPeerAvailable, err)

	cctx, cancel := context.WithCancel(ctx)
	cancel()

	p, err = b.Get(cctx, balancer.GetOptions{})

	assert.Nil(t, p)
	assert.Equal(t, err, context.Canceled)

	err = b.Close()
	assert.Nil(t, p)

	p, err = b.Get(ctx, balancer.GetOptions{})

	assert.Nil(t, p)
	assert.Equal(t, err, context.Canceled)
}

func TestBalanceWithPerrs(t *testing.T) {
	ctx := context.Background()
	b := NewBalancer(
		static.NewResolverFromStrings([]string{"localhost:0", "localhost:1"}),
	)

	err := b.Open(ctx)
	assert.Nil(t, err)

	p, err := b.Get(ctx, balancer.GetOptions{})

	assert.Nil(t, err)
	assert.Equal(t, "localhost:0", p.Addr())

	p, err = b.Get(ctx, balancer.GetOptions{})

	assert.Nil(t, err)
	assert.Equal(t, "localhost:1", p.Addr())

	p, err = b.Get(ctx, balancer.GetOptions{})

	assert.Nil(t, err)
	assert.Equal(t, "localhost:0", p.Addr())

	b.Close()
}
