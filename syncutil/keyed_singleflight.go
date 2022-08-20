package syncutil

import (
	"context"
	"sync"

	"github.com/upfluence/pkg/group"
)

type KeyedSingleflight[K comparable, V any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once

	mu  sync.Mutex
	sfs map[K]*Singleflight[map[K]V]
}

func (ksf *KeyedSingleflight[K, V]) init() {
	ksf.once.Do(func() {
		ksf.ctx, ksf.cancel = context.WithCancel(context.Background())
		ksf.sfs = make(map[K]*Singleflight[map[K]V])
	})
}

func (ksf *KeyedSingleflight[K, V]) releaseKeys(keys []K) {
	ksf.mu.Lock()
	defer ksf.mu.Unlock()

	for _, key := range keys {
		delete(ksf.sfs, key)
	}
}

func (ksf *KeyedSingleflight[K, V]) prepare(keys []K) executors[K, V] {
	var (
		newSingleflight       *Singleflight[map[K]V]
		keysToExecute         []K
		existingSingleflights = make(map[*Singleflight[map[K]V]][]K)
	)

	ksf.init()

	ksf.mu.Lock()

	for _, key := range keys {
		if sf, ok := ksf.sfs[key]; ok {
			existingSingleflights[sf] = append(existingSingleflights[sf], key)
			continue
		}

		keysToExecute = append(keysToExecute, key)
	}

	if len(keysToExecute) > 0 {
		newSingleflight = &Singleflight[map[K]V]{ctx: ksf.ctx, cancel: ksf.cancel}

		for _, k := range keysToExecute {
			ksf.sfs[k] = newSingleflight
		}
	}

	ksf.mu.Unlock()

	executors := make(executors[K, V], 0, len(existingSingleflights)+1)

	for sf, ks := range existingSingleflights {
		executors = append(
			executors,
			executor[K, V]{singleflight: sf, keys: ks, ksf: ksf},
		)
	}

	if newSingleflight != nil {
		executors = append(
			executors,
			executor[K, V]{
				singleflight: newSingleflight,
				keys:         keysToExecute,
				ksf:          ksf,
			},
		)
	}

	return executors
}

func (ksf *KeyedSingleflight[K, V]) Do(ctx context.Context, keys []K, fn func(context.Context, []K) (map[K]V, error)) (map[K]V, error) {
	return ksf.prepare(keys).execute(ctx, fn)
}

func (ksf *KeyedSingleflight[K, V]) Close() error {
	ksf.init()
	ksf.cancel()

	for _, sf := range ksf.sfs {
		sf.wg.Wait()
	}

	return nil
}

type executor[K comparable, V any] struct {
	keys         []K
	singleflight *Singleflight[map[K]V]

	ksf *KeyedSingleflight[K, V]
}

func (e executor[K, V]) execute(ctx context.Context, fn func(context.Context, []K) (map[K]V, error)) (bool, map[K]V, error) {
	ok, vs, err := e.singleflight.Do(
		ctx,
		func(ctx context.Context) (map[K]V, error) {
			res, err := fn(ctx, e.keys)
			e.ksf.releaseKeys(e.keys)
			return res, err
		},
	)

	res := make(map[K]V, len(e.keys))

	for _, key := range e.keys {
		if v, ok := vs[key]; ok {
			res[key] = v
		}
	}

	return ok, res, err
}

type executors[K comparable, V any] []executor[K, V]

func (es executors[K, V]) execute(ctx context.Context, fn func(context.Context, []K) (map[K]V, error)) (map[K]V, error) {
	switch len(es) {
	case 0:
		return nil, nil
	case 1:
		_, vs, err := es[0].execute(ctx, fn)
		return vs, err
	}

	tg := group.TypedGroup[map[K]V]{
		Group: group.ErrorGroup(ctx),
		Value: make(map[K]V),
	}

	for _, e := range es {
		e := e

		tg.Do(func(ctx context.Context) (func(*map[K]V), error) {
			_, vs, err := e.execute(ctx, fn)

			if err != nil {
				return nil, err
			}

			return func(res *map[K]V) {
				for k, v := range vs {
					(*res)[k] = v
				}
			}, nil
		})
	}

	return tg.Wait()
}

func DoOne[K comparable, V any](ctx context.Context, ksf *KeyedSingleflight[K, V], key K, fn func(context.Context) (V, error)) (bool, V, error) {
	ok, vs, err := ksf.prepare([]K{key})[0].execute(
		ctx,
		func(ctx context.Context, _ []K) (map[K]V, error) {
			v, err := fn(ctx)

			if err != nil {
				return nil, err
			}

			return map[K]V{key: v}, nil
		},
	)

	if err != nil {
		var zero V
		return ok, zero, err
	}

	return ok, vs[key], nil
}
