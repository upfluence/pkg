package resolver

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/closer"
	"github.com/upfluence/pkg/v2/discovery/peer"
	"github.com/upfluence/pkg/v2/log"
)

type Puller[T peer.Peer] struct {
	Resolver   Resolver[T]
	UpdateFunc func(Update[T])
	Monitor    closer.Monitor
	NoWait     bool

	openErr   error
	openOnce  sync.Once
	closeOnce sync.Once
	opened    atomic.Bool
}

func NewPuller[T peer.Peer](r Resolver[T], fn func(Update[T])) (*Puller[T], func() error) {
	var p = &Puller[T]{
		Resolver:   r,
		UpdateFunc: fn,
	}

	return p, p.Close
}

func (p *Puller[T]) Close() error {
	var err error

	p.closeOnce.Do(func() {
		p.opened.Store(false)
		err = errors.Combine(p.Monitor.Close(), p.Resolver.Close())
	})

	return err
}

func (p *Puller[T]) IsOpen() bool {
	return p.opened.Load()
}

func (p *Puller[T]) String() string {
	return fmt.Sprintf("%v", p.Resolver)
}

func (p *Puller[T]) Open(ctx context.Context) error {
	p.openOnce.Do(func() {
		p.openErr = p.Resolver.Open(ctx)

		if p.openErr == nil {
			p.opened.Store(true)
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
		err = nil
		w = p.Resolver.Resolve()

		for err == nil {
			u, err = w.Next(ctx, ResolveOptions{NoWait: noWait})

			if err == nil {
				noWait = false

				p.UpdateFunc(u)

				continue
			}

			if errors.Is(err, ErrNoUpdates) {
				// No update available right now; reset noWait and keep going
				// without calling UpdateFunc with a meaningless empty update.
				noWait = false

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
