package resolver

import (
	"context"
	"errors"
	"io"

	"github.com/upfluence/pkg/discovery/peer"
)

type Update struct {
	Additions []peer.Peer
	Deletions []peer.Peer
}

type Builder interface {
	Build(string) Resolver
}

type BuilderFunc func(string) Resolver

func (fn BuilderFunc) Build(k string) Resolver { return fn(k) }

var (
	ErrNoUpdates = errors.New("discovery/resolver: No update available")
	ErrClose     = errors.New("discovery/resolver: Resolver close")
)

type ResolveOptions struct {
	NoWait bool
}

type Resolver interface {
	io.Closer

	Open(context.Context) error

	Resolve() Watcher
}

type Watcher interface {
	io.Closer

	Next(context.Context, ResolveOptions) (Update, error)
}
