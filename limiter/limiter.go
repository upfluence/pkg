package limiter

import (
	"context"

	"github.com/upfluence/errors"
)

var ErrLimited = errors.New("limiter: Limited")

type DoneFunc func()

type AllowOptions struct {
	NoWait bool
	N      int
}

type Limiter interface {
	Allow(context.Context, AllowOptions) (DoneFunc, error)
}
