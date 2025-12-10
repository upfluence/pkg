package group

import "context"

type throttledGroup struct {
	Group

	ch chan struct{}
}

func (tg *throttledGroup) Do(r Runner) {
	tg.Group.Do(wrapRunner(r, tg.ch))
}

// SharedThrottledGroup wraps a Group with a shared concurrency limiter.
// The provided channel controls how many goroutines can execute concurrently.
// This allows multiple throttled groups to share the same concurrency limit.
//
// The channel should be buffered, where the buffer size determines the maximum
// number of concurrent executions across all groups sharing the channel.
//
// Example:
//
//	// Create a shared limiter for 5 concurrent operations
//	limiter := make(chan struct{}, 5)
//
//	// Two groups sharing the same limit
//	g1 := group.SharedThrottledGroup(group.WaitGroup(ctx), limiter)
//	g2 := group.SharedThrottledGroup(group.WaitGroup(ctx), limiter)
//
//	// Both groups combined won't exceed 5 concurrent operations
//	for i := 0; i < 10; i++ {
//		g1.Do(func(ctx context.Context) error {
//			return processTask(ctx)
//		})
//	}
//
//	for i := 0; i < 10; i++ {
//		g2.Do(func(ctx context.Context) error {
//			return processTask(ctx)
//		})
//	}
func SharedThrottledGroup(g Group, ch chan struct{}) Group {
	return &throttledGroup{Group: g, ch: ch}
}

// ThrottledGroup wraps a Group with a concurrency limiter that restricts
// the maximum number of goroutines executing concurrently.
//
// This is useful for rate limiting operations against external services,
// managing resource usage, or preventing overwhelming a system with too
// many concurrent operations.
//
// Example:
//
//	g := group.WaitGroup(ctx)
//	throttled := group.ThrottledGroup(g, 10) // max 10 concurrent
//
//	// Process 100 items with only 10 running at a time
//	for i := 0; i < 100; i++ {
//		throttled.Do(func(ctx context.Context) error {
//			return expensiveOperation(ctx, i)
//		})
//	}
//
//	return throttled.Wait()
func ThrottledGroup(g Group, cap int) Group {
	return SharedThrottledGroup(g, make(chan struct{}, cap))
}

func wrapRunner(r Runner, ch chan struct{}) Runner {
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- struct{}{}:
		}

		err := r(ctx)

		select {
		case <-ch:
		default:
		}

		return err
	}
}
