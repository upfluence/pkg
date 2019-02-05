package group

import (
	"context"
)

type Runner func(context.Context) error

type Group interface {
	Do(Runner)

	Wait() error
}
