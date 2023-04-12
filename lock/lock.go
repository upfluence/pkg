package lock

import (
	"context"
	"time"

	"github.com/upfluence/errors"
)

var (
	ErrPastTime        = errors.New("Deadline is in the past")
	ErrAlreadyAcquired = errors.New("Lock already acquired")
	ErrLeaseNotFound   = errors.New("Lease not found")
)

type OpType uint8

const (
	Write OpType = iota
	Read
)

type AcquireOptions struct {
	NoWait   bool
	Deadline time.Time
	Op       OpType
}

type LockManager interface {
	Lock(string) Lock
}

type Lock interface {
	Acquire(context.Context, AcquireOptions) (Lease, error)
}

type Lease interface {
	Release(context.Context) error
	KeepAlive(context.Context, time.Time) error

	Done() <-chan struct{}
	Err() error
	Deadline() time.Time
}
