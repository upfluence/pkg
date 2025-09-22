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
	b := NewBalancer(&static.Resolver[static.Peer]{})

	p, done, err := b.Get(ctx, balancer.GetOptions{NoWait: true})

	assert.Empty(t, p.Addr())
	assert.Nil(t, done)
	assert.Equal(t, balancer.ErrNoPeerAvailable, err)

	cctx, cancel := context.WithCancel(ctx)
	cancel()

	p, done, err = b.Get(cctx, balancer.GetOptions{})

	assert.Empty(t, p.Addr())
	assert.Nil(t, done)
	assert.Equal(t, err, context.Canceled)

	err = b.Close()
	assert.NoError(t, err)

	p, done, err = b.Get(ctx, balancer.GetOptions{})

	assert.Empty(t, p.Addr())
	assert.Nil(t, done)
	assert.Equal(t, err, context.Canceled)
}

func TestBalanceWithPerrs(t *testing.T) {
	ctx := context.Background()
	b := NewBalancer(
		static.NewResolverFromStrings([]string{"localhost:0", "localhost:1"}),
	)

	err := b.Open(ctx)
	assert.Nil(t, err)

	p, done, err := b.Get(ctx, balancer.GetOptions{})
	done(nil)

	assert.Nil(t, err)
	assert.Equal(t, "localhost:0", p.Addr())

	p, done, err = b.Get(ctx, balancer.GetOptions{})
	done(nil)

	assert.Nil(t, err)
	assert.Equal(t, "localhost:1", p.Addr())

	p, done, err = b.Get(ctx, balancer.GetOptions{})
	done(nil)

	assert.Nil(t, err)
	assert.Equal(t, "localhost:0", p.Addr())

	b.Close()
}
