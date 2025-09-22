package resolver

import (
	"context"
	"io"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/discovery/peer"
)

type Update[T peer.Peer] struct {
	Additions []T
	Deletions []T
}

type Builder[T peer.Peer] interface {
	Build(string) Resolver[T]
}

type BuilderFunc[T peer.Peer] func(string) Resolver[T]

func (fn BuilderFunc[T]) Build(k string) Resolver[T] { return fn(k) }

var (
	ErrNoUpdates = errors.New("discovery/resolver: No update available")
	ErrClose     = errors.New("discovery/resolver: Resolver close")
)

type ResolveOptions struct {
	NoWait bool
}

type Resolver[T peer.Peer] interface {
	io.Closer

	Open(context.Context) error

	Resolve() Watcher[T]
}

type Watcher[T peer.Peer] interface {
	io.Closer

	Next(context.Context, ResolveOptions) (Update[T], error)
}
