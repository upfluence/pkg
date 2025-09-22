package limiter

import (
	"context"

	"github.com/upfluence/pkg/v2/group"
	"github.com/upfluence/pkg/v2/limiter"
)

type Group struct {
	g group.Group
	l limiter.Limiter

	noWait bool
}

func WrapGroup(g group.Group, l limiter.Limiter) *Group {
	return &Group{g: g, l: l}
}

func (g *Group) Do(r group.Runner) { g.g.Do(wrapRunner(r, g.l, g.noWait)) }
func (g *Group) Wait() error       { return g.g.Wait() }

func wrapRunner(r group.Runner, l limiter.Limiter, noWait bool) group.Runner {
	return func(ctx context.Context) error {
		var done, err = l.Allow(
			ctx,
			limiter.AllowOptions{N: 1, NoWait: noWait},
		)

		if err != nil {
			return err
		}

		err = r(ctx)
		done()

		return err
	}
}
