package static

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/discovery/resolver"
)

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
