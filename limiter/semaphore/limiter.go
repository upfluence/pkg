package semaphore

import (
	"context"
	"fmt"
	"sync"

	"github.com/upfluence/pkg/limiter"
)

type Limiter struct {
	cond *sync.Cond
	mu   sync.Mutex

	remaining int
	size      int
}

func NewLimiter(size int) *Limiter {
	var l = Limiter{remaining: size, size: size}

	l.cond = sync.NewCond(&l.mu)

	return &l
}

func (l *Limiter) String() string {
	return fmt.Sprintf("limiter/semaphore: [size: %d, remaining: %d]", l.size, l.remaining)
}

func (l *Limiter) release(n int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.remaining += n
	l.cond.Broadcast()
}

func (l *Limiter) Update(size int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.remaining += size - l.size
	l.size = size

	if l.remaining > 0 {
		l.cond.Broadcast()
	}
}

func (l *Limiter) Allow(ctx context.Context, opts limiter.AllowOptions) (limiter.DoneFunc, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	donefn := func() { l.release(opts.N) }

	if opts.NoWait {
		if l.remaining < opts.N {
			return nil, limiter.ErrLimited
		}

		l.remaining -= opts.N
		return donefn, nil
	}

	var (
		wg sync.WaitGroup

		done = make(chan struct{})
	)

	wg.Add(1)

	go func() {
		select {
		case <-ctx.Done():
			l.cond.Broadcast()
		case <-done:
		}

		wg.Done()
	}()

	for l.remaining < opts.N {
		select {
		case <-ctx.Done():
			close(done)
			wg.Wait()

			return nil, ctx.Err()
		default:
		}

		l.cond.Wait()
	}

	close(done)
	wg.Wait()

	return donefn, nil
}
