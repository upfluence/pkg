package syncutil

import (
	"context"
	"sync"
)

type result[T any] struct {
	res T
	err error
}

type Singleflight[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
	wg     sync.WaitGroup

	mu  sync.Mutex
	chs []chan<- result[T]
}

func (sf *Singleflight[T]) init() {
	sf.once.Do(func() {
		sf.ctx, sf.cancel = context.WithCancel(context.Background())
	})
}

func (sf *Singleflight[T]) Do(ctx context.Context, fn func(context.Context) (T, error)) (bool, T, error) {
	sf.init()

	ch := make(chan result[T], 1)
	executor := false

	sf.mu.Lock()
	sf.chs = append(sf.chs, ch)
	executor = len(sf.chs) == 1

	if executor {
		sf.wg.Add(1)

		go func() {
			val, err := fn(sf.ctx)
			res := result[T]{res: val, err: err}

			sf.mu.Lock()

			for _, ch := range sf.chs {
				ch <- res
				close(ch)
			}

			sf.chs = nil

			sf.mu.Unlock()

			sf.wg.Done()
		}()
	}

	sf.mu.Unlock()

	select {
	case <-ctx.Done():
		var zero T
		return executor, zero, ctx.Err()
	case res := <-ch:
		return executor, res.res, res.err
	}
}

func (sf *Singleflight[T]) Close() error {
	sf.init()
	sf.cancel()
	sf.wg.Wait()
	return nil
}
