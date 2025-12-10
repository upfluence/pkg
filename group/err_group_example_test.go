package group_test

import (
	"context"
	"fmt"
	"time"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/group"
)

func ExampleExitGroup() {
	ctx := context.Background()
	g := group.ExitGroup(ctx)

	// First task completes quickly
	g.Do(func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		fmt.Println("Task 1 completed")
		return nil
	})

	// Second task would take longer but gets cancelled
	g.Do(func(ctx context.Context) error {
		select {
		case <-time.After(1 * time.Second):
			fmt.Println("Task 2 completed")
			return nil
		case <-ctx.Done():
			fmt.Println("Task 2 cancelled")
			return ctx.Err()
		}
	})

	// Wait returns when first task completes
	if err := g.Wait(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Output:
	// Task 1 completed
	// Task 2 cancelled
}

func ExampleErrorGroup() {
	ctx := context.Background()
	g := group.ErrorGroup(ctx)

	// Task that returns an error
	g.Do(func(ctx context.Context) error {
		return errors.New("processing failed")
	})

	// This task gets cancelled due to error above
	g.Do(func(ctx context.Context) error {
		select {
		case <-time.After(1 * time.Second):
			return fmt.Errorf("should not complete")
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	// Wait returns the first error encountered
	// The context cancellation error is ignored
	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}

	// Output: processing failed
}
