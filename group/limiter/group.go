// Package limiter provides integration between the group package and rate limiters,
// allowing concurrent goroutine execution to be controlled by sophisticated rate
// limiting strategies.
//
// This package wraps any group.Group with a limiter.Limiter, enabling fine-grained
// control over resource consumption through token bucket algorithms, sliding windows,
// or other rate limiting mechanisms.
//
// Common Pattern:
//
//	func ProcessWithRateLimit(ctx context.Context, items []Item) error {
//		// Create a rate limiter (e.g., 100 requests per second)
//		rateLimiter := rate.NewLimiter(100, 10)
//
//		// Wrap a group with the limiter
//		g := limiter.WrapGroup(
//			group.ErrorGroup(ctx),
//			rateLimiter,
//		)
//
//		for _, item := range items {
//			item := item
//			g.Do(func(ctx context.Context) error {
//				// This will wait for rate limiter permission
//				return processItem(ctx, item)
//			})
//		}
//
//		return g.Wait()
//	}
package limiter

import (
	"context"

	"github.com/upfluence/pkg/v2/group"
	"github.com/upfluence/pkg/v2/limiter"
)

// Group wraps a group.Group with a limiter.Limiter to provide rate-limited
// concurrent execution. Each goroutine scheduled via Do will acquire a token
// from the limiter before executing.
//
// The noWait field controls whether runners should wait for limiter availability
// or fail immediately if no tokens are available.
//
// Example:
//
//	rateLimiter := rate.NewLimiter(10, 5) // 10 ops/sec, burst of 5
//	g := limiter.WrapGroup(group.WaitGroup(ctx), rateLimiter)
//
//	for i := 0; i < 100; i++ {
//		i := i
//		g.Do(func(ctx context.Context) error {
//			// Rate limited to 10/sec
//			return apiCall(ctx, i)
//		})
//	}
//
//	return g.Wait()
type Group struct {
	g group.Group
	l limiter.Limiter

	noWait bool
}

// WrapGroup creates a new Group that wraps the provided group.Group with
// rate limiting from the provided limiter.Limiter.
//
// Each runner scheduled via Do will acquire tokens from the limiter before
// executing. By default, runners will wait for tokens to become available.
// The noWait behavior can be configured by modifying the returned Group.
//
// This is useful for controlling resource usage when making external API calls,
// database queries, or any rate-sensitive operations.
//
// Example:
//
//	// Limit API calls to 50 per second
//	rateLimiter := rate.NewLimiter(rate.Limit(50), 10)
//
//	g := limiter.WrapGroup(
//		group.ErrorGroup(ctx),
//		rateLimiter,
//	)
//
//	for _, userID := range userIDs {
//		g.Do(func(ctx context.Context) error {
//			// Automatically rate limited
//			return fetchUserProfile(ctx, userID)
//		})
//	}
//
//	if err := g.Wait(); err != nil {
//		return fmt.Errorf("failed to fetch profiles: %w", err)
//	}
func WrapGroup(g group.Group, l limiter.Limiter) *Group {
	return &Group{g: g, l: l}
}

// Do schedules a runner to execute concurrently with rate limiting.
// Before the runner executes, it acquires a token from the limiter.
// The token is automatically released when the runner completes.
//
// If noWait is false (default), Do will wait for a token to become available.
// If noWait is true, Do will fail immediately if no tokens are available.
//
// Example:
//
//	g.Do(func(ctx context.Context) error {
//		// This waits for rate limiter permission before executing
//		resp, err := http.Post(apiURL, contentType, body)
//		if err != nil {
//			return fmt.Errorf("API call failed: %w", err)
//		}
//		defer resp.Body.Close()
//		return processResponse(resp)
//	})
func (g *Group) Do(r group.Runner) { g.g.Do(wrapRunner(r, g.l, g.noWait)) }

func (g *Group) Wait() error { return g.g.Wait() }

func wrapRunner(r group.Runner, l limiter.Limiter, noWait bool) group.Runner {
	return func(ctx context.Context) error {
		var done, err = l.Allow(
			ctx,
			limiter.AllowOptions{N: 1, NoWait: noWait},
		)

		if err != nil {
			return err
		}

		err = r(ctx)
		done()

		return err
	}
}
