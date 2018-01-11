package balancer

import (
	"context"
	"errors"

	"github.com/upfluence/pkg/discovery/peer"
)

var ErrNoPeerAvailable = errors.New("balancer: No peer available")

type BalancerGetOptions struct {
	NoWait bool
}

type Balancer interface {
	Open(context.Context) error
	IsOpen() bool
	Close() error

	// Up(peer.Peer) func(error)
	Get(context.Context, BalancerGetOptions) (*peer.Peer, error)
}
