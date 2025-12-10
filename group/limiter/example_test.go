package limiter_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/upfluence/pkg/v2/group"
	"github.com/upfluence/pkg/v2/group/limiter"
	"github.com/upfluence/pkg/v2/limiter/rate"
)

func Example() {
	// 10 ops/sec, burst of 5
	rateLimiter := rate.NewLimiter(rate.Config{
		Baseline: 10,
		Period:   time.Second,
		Burst:    5,
	})

	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second+time.Millisecond))

	g := limiter.WrapGroup(group.WaitGroup(ctx), rateLimiter)

	var count atomic.Int32

	for i := 0; i < 100; i++ {
		g.Do(func(ctx context.Context) error {
			count.Add(1)
			return nil
		})
	}

	_ = g.Wait()
	fmt.Println("Call count:", count.Load())
	// Output: Call count: 15
}
