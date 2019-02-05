package group

import (
	"context"
	"testing"
	"time"
)

func TestThrottledGroup(t *testing.T) {
	g := ThrottledGroup(WaitGroup(context.Background()), 1)

	t0 := time.Now()

	g.Do(func(context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	g.Do(func(context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if err := g.Wait(); err != nil {
		t.Errorf("Wait error = %v; wants nil", err)
	}

	if time.Since(t0) < 20*time.Millisecond {
		t.Errorf("ThrottledGroup as parralized the process")
	}
}
