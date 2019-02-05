package group

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestErrorGroup(t *testing.T) {
	var errMock = errors.New("err1")

	g := ErrorGroup(context.Background())

	g.Do(func(context.Context) error { return nil })
	time.Sleep(10 * time.Millisecond)
	g.Do(func(context.Context) error { return errMock })

	if err := g.Wait(); err != errMock {
		t.Errorf("Wait() = %v, wanted %v", err, errMock)
	}
}

func TestExitGroup(t *testing.T) {
	var errMock = errors.New("err1")

	g := ExitGroup(context.Background())

	g.Do(func(context.Context) error { return nil })
	time.Sleep(10 * time.Millisecond)
	g.Do(func(context.Context) error { return errMock })

	if err := g.Wait(); err != nil {
		t.Errorf("Wait() = %v, wanted nil", err)
	}
}
