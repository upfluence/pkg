package group_test

import (
	"context"
	"fmt"
	"time"

	"github.com/upfluence/pkg/v2/group"
)

func ExampleThrottledGroup() {
	ctx, cancel := context.WithCancel(context.Background())
	g := group.ThrottledGroup(group.WaitGroup(ctx), 1)

	g.Do(func(context.Context) error {
		time.Sleep(10 * time.Millisecond)
		fmt.Println("done")
		cancel()

		return nil
	})

	g.Do(func(context.Context) error {
		time.Sleep(10 * time.Millisecond)
		fmt.Println("done")
		cancel()

		return nil
	})

	// Expecting only one "done" because of the throttling the second task should not start before the first one is done.
	// And since we cancel the context in the first task, the second one should not run at all.
	_ = g.Wait() // nolint:errcheck
	// Output: done
}
