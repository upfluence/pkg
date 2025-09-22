package group

import (
	"context"
	"testing"
)

func TestTypedGroup(t *testing.T) {
	readyc := make(chan struct{})
	g := TypedGroup[int]{Group: ErrorGroup(context.Background())}

	for i := 0; i < 5; i++ {
		g.Do(func(context.Context) (func(*int), error) {
			<-readyc
			return func(v *int) { *v++ }, nil
		})
	}

	close(readyc)

	res, err := g.Wait()

	if res != 5 {
		t.Errorf("Wait() = (%v, _), wanted 5", res)
	}

	if err != nil {
		t.Errorf("Wait() = (_, %v), wanted nil", err)
	}
}
