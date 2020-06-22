package syncutil

import (
	"context"
	"sync"
)

type Singleflight struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
	wg     sync.WaitGroup

	mu  sync.Mutex
	chs []chan<- error
}

func (sf *Singleflight) init() {
	sf.once.Do(func() {
		sf.ctx, sf.cancel = context.WithCancel(context.Background())
	})
}

func (sf *Singleflight) Do(ctx context.Context, fn func(context.Context) error) (bool, error) {
	sf.init()

	ch := make(chan error, 1)
	executor := false

	sf.mu.Lock()
	sf.chs = append(sf.chs, ch)
	executor = len(sf.chs) == 1

	if executor {
		sf.wg.Add(1)

		go func() {
			err := fn(sf.ctx)

			sf.mu.Lock()

			for _, ch := range sf.chs {
				ch <- err
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
		return executor, ctx.Err()
	case err := <-ch:
		return executor, err
	}
}

func (sf *Singleflight) Close() error {
	sf.init()
	sf.cancel()
	sf.wg.Wait()
	return nil
}
