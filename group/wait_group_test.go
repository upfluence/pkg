package group

import (
	"context"
	"errors"
	"testing"

	"github.com/upfluence/pkg/multierror"
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

	merr, ok := err.(multierror.MultiError)

	if !ok || len(merr.Errors()) != 2 {
		t.Errorf("Wait() = %v, wanted MultiError([%v, %v])", err, err1, err2)
	}
}
