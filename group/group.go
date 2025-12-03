// Package group provides utilities for managing concurrent goroutine execution
// with various synchronization and error handling strategies.
//
// The package offers several implementations of the Group interface, each with
// different characteristics for handling concurrency, errors, and resource limits.
//
// Common Pattern:
//
//	func ProcessItems(ctx context.Context, items []Item) error {
//		// Create an error group that stops on first error
//		g := group.ErrorGroup(ctx)
//
//		for _, item := range items {
//			g.Do(func(ctx context.Context) error {
//				return processItem(ctx, item)
//			})
//		}
//
//		// Wait for all goroutines to complete
//		return g.Wait()
//	}
//
// For rate-limited concurrent execution:
//
//	func ProcessWithLimit(ctx context.Context, items []Item) error {
//		g := group.WaitGroup(ctx)
//		throttled := group.ThrottledGroup(g, 10) // max 10 concurrent
//
//		for _, item := range items {
//			item := item
//			throttled.Do(func(ctx context.Context) error {
//				return processItem(ctx, item)
//			})
//		}
//
//		return throttled.Wait()
//	}
package group

import (
	"context"
)

// Runner is a function type that executes work in a goroutine context.
// It receives a context.Context for cancellation and deadline support,
// and returns an error if the operation fails.
//
// Example:
//
//	var runner group.Runner = func(ctx context.Context) error {
//		resp, err := http.Get("https://api.example.com/data")
//		if err != nil {
//			return fmt.Errorf("fetch failed: %w", err)
//		}
//		defer resp.Body.Close()
//		return processResponse(resp)
//	}
type Runner func(context.Context) error

// Group is the core interface for managing concurrent goroutine execution.
// Implementations provide different strategies for error handling and
// synchronization.
//
// The Do method schedules a Runner to execute concurrently, and Wait blocks
// until all scheduled runners complete, returning any errors according to
// the implementation's strategy.
//
// Example:
//
//	g := group.WaitGroup(context.Background())
//
//	g.Do(func(ctx context.Context) error {
//		return task1(ctx)
//	})
//
//	g.Do(func(ctx context.Context) error {
//		return task2(ctx)
//	})
//
//	if err := g.Wait(); err != nil {
//		log.Fatal(err)
//	}
type Group interface {
	// Do schedules a Runner to execute concurrently.
	// The behavior when called after Wait or when the context is cancelled
	// depends on the implementation.
	//
	// Example:
	//
	//	g.Do(func(ctx context.Context) error {
	//		select {
	//		case <-ctx.Done():
	//			return ctx.Err()
	//		case <-time.After(time.Second):
	//			return performWork()
	//		}
	//	})
	Do(Runner)

	// Wait blocks until all scheduled runners complete and returns any
	// errors according to the implementation's error handling strategy.
	// After Wait returns, the Group should not be reused.
	//
	// Example:
	//
	//	if err := g.Wait(); err != nil {
	//		return fmt.Errorf("group execution failed: %w", err)
	//	}
	Wait() error
}
