package resolver

import (
	"context"
	"fmt"
	"sync"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/closer"
	"github.com/upfluence/pkg/log"
)

type Puller struct {
	Resolver   Resolver
	UpdateFunc func(Update)
	Monitor    closer.Monitor
	NoWait     bool

	openErr  error
	openOnce sync.Once
}

func NewPuller(r Resolver, fn func(Update)) (*Puller, func()) {
	var p = &Puller{
		Resolver:   r,
		UpdateFunc: fn,
	}

	return p, func() { p.Close() }
}

func (p *Puller) Close() error {
	return errors.Combine(p.Monitor.Close(), p.Resolver.Close())
}

func (p *Puller) IsOpen() bool {
	return p.openErr == nil && p.openOnce != sync.Once{}
}

func (p *Puller) String() string {
	return fmt.Sprintf("%v", p.Resolver)
}

func (p *Puller) Open(ctx context.Context) error {
	p.openOnce.Do(func() {
		p.openErr = p.Resolver.Open(ctx)

		if p.openErr == nil {
			p.Monitor.Run(p.pull)
		}
	})

	return p.openErr
}

func (p *Puller) pull(ctx context.Context) {
	var (
		u   Update
		err error
		w   Watcher

		noWait = p.NoWait
	)

	for {
		w = p.Resolver.Resolve()

		for err == nil {
			u, err = w.Next(ctx, ResolveOptions{NoWait: noWait})

			if err == nil || err == ErrNoUpdates {
				noWait = false
				p.UpdateFunc(u)
				continue
			}

			if ctx.Err() == nil {
				log.WithError(err).Error("resolving failed")
			}

			w.Close()
		}

		if ctx.Err() != nil {
			return
		}
	}
}
