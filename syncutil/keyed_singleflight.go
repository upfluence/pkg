package syncutil

import (
	"context"
	"sync"
)

type KeyedSingleflight[K comparable] struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once

	mu  sync.Mutex
	sfs map[K]*Singleflight
}

func (ksf *KeyedSingleflight[K]) init() {
	ksf.once.Do(func() {
		ksf.ctx, ksf.cancel = context.WithCancel(context.Background())
		ksf.sfs = make(map[K]*Singleflight)
	})
}

func (ksf *KeyedSingleflight[K]) Do(ctx context.Context, key K, fn func(context.Context) error) (bool, error) {
	ksf.init()

	ksf.mu.Lock()

	sf, ok := ksf.sfs[key]

	if !ok {
		sf = &Singleflight{ctx: ksf.ctx, cancel: ksf.cancel}
		ksf.sfs[key] = sf
	}

	ksf.mu.Unlock()

	return sf.Do(
		ctx,
		func(ctx context.Context) error {
			err := fn(ctx)

			ksf.mu.Lock()
			delete(ksf.sfs, key)
			ksf.mu.Unlock()

			return err
		},
	)
}

func (ksf *KeyedSingleflight[K]) Close() error {
	ksf.init()
	ksf.cancel()

	for _, sf := range ksf.sfs {
		sf.wg.Wait()
	}

	return nil
}
