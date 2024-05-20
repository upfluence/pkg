package resolver

import (
	"context"
	"fmt"
	"sync"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/closer"
	"github.com/upfluence/pkg/discovery/peer"
	"github.com/upfluence/pkg/log"
)

type Puller[T peer.Peer] struct {
	Resolver   Resolver[T]
	UpdateFunc func(Update[T])
	Monitor    closer.Monitor
	NoWait     bool

	openErr  error
	openOnce sync.Once
}

func NewPuller[T peer.Peer](r Resolver[T], fn func(Update[T])) (*Puller[T], func()) {
	var p = &Puller[T]{
		Resolver:   r,
		UpdateFunc: fn,
	}

	return p, func() { p.Close() }
}

func (p *Puller[T]) Close() error {
	return errors.Combine(p.Monitor.Close(), p.Resolver.Close())
}

func (p *Puller[T]) IsOpen() bool {
	return p.openErr == nil && p.openOnce != sync.Once{}
}

func (p *Puller[T]) String() string {
	return fmt.Sprintf("%v", p.Resolver)
}

func (p *Puller[T]) Open(ctx context.Context) error {
	p.openOnce.Do(func() {
		p.openErr = p.Resolver.Open(ctx)

		if p.openErr == nil {
			p.Monitor.Run(p.pull)
		}
	})

	return p.openErr
}

func (p *Puller[T]) pull(ctx context.Context) {
	var (
		u   Update[T]
		err error
		w   Watcher[T]

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
