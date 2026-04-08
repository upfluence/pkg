package group

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/errors"
	"github.com/upfluence/errors/multi"
)

func TestWaitGroup(t *testing.T) {
	var (
		err1 = errors.New("err1")
		err2 = errors.New("err2")
	)

	g := WaitGroup(context.Background())

	g.Do(func(context.Context) error { return err1 })
	g.Do(func(context.Context) error { return err2 })

	err := g.Wait()

	if errs := multi.ExtractErrors(err); len(errs) != 2 {
		t.Errorf("Wait() = %v, wanted MultiError([%v, %v])", err, err1, err2)
	}
}

func TestWaitGroupPreCanceled(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	errPreCanceled := errors.New("pre-canceled")

	cancel(errPreCanceled)

	for _, tt := range []struct {
		name string
		gfn  func(context.Context) Group
	}{
		{name: "WaitGroup", gfn: WaitGroup},
		{name: "ErrorGroup", gfn: ErrorGroup},
		{name: "ExitGroup", gfn: ExitGroup},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.gfn(ctx)

			g.Do(func(context.Context) error { return errors.New("should not run") })

			err := g.Wait()

			assert.ErrorIs(t, err, errPreCanceled)
		})
	}
}
