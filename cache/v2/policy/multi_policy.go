package policy

import (
	"context"
	"sync"

	"github.com/upfluence/errors"
)

func CombinePolicies[K comparable](ps ...EvictionPolicy[K]) EvictionPolicy[K] {
	switch len(ps) {
	case 0:
		return &NopPolicy[K]{}
	case 1:
		return ps[0]
	}

	l := ps[0]

	for i := 1; i < len(ps); i++ {
		l = newMultiPolicty(l, ps[i])
	}

	return l
}

func newMultiPolicty[K comparable](l, r EvictionPolicy[K]) *multiPolicy[K] {
	mp := multiPolicy[K]{
		l:  l,
		r:  r,
		ch: make(chan K),
	}

	mp.wg.Add(1)
	mp.ctx, mp.cancel = context.WithCancel(context.Background())

	go mp.pull()

	return &mp
}

type multiPolicy[K comparable] struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan K

	l, r EvictionPolicy[K]
}

func (mp *multiPolicy[K]) pull() {
	defer mp.wg.Done()

	for {
		select {
		case <-mp.ctx.Done():
			return
		case k := <-mp.l.C():
			mp.r.Op(k, Evict)
			mp.ch <- k
		case k := <-mp.r.C():
			mp.l.Op(k, Evict)
			mp.ch <- k
		}
	}
}

func (mp *multiPolicy[K]) C() <-chan K {
	return mp.ch
}

func (mp *multiPolicy[K]) Op(k K, op OpType) error {
	return errors.Combine(mp.l.Op(k, op), mp.r.Op(k, op))
}

func (mp *multiPolicy[K]) Close() error {
	mp.cancel()
	mp.wg.Wait()
	close(mp.ch)

	return errors.Combine(mp.l.Close(), mp.r.Close())
}
