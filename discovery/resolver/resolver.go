package resolver

import (
	"context"

	"github.com/upfluence/pkg/discovery/peer"
)

type Update struct {
	Additions []*peer.Peer
	Deletions []*peer.Peer
}

type Target struct {
	Identifier, Environment string
}

type Builder interface {
	Build(Target) (Resolver, error)
}

type Resolver interface {
	Open(context.Context) error
	Close() error

	Resolve(context.Context) (<-chan Update, error)
}
