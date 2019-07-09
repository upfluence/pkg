package policy

import (
	"context"
	"sync"

	"github.com/upfluence/pkg/multierror"
)

func CombinePolicies(ps ...EvictionPolicy) EvictionPolicy {
	switch len(ps) {
	case 0:
		return &NopPolicy{}
	case 1:
		return ps[0]
	}

	l := ps[0]

	for i := 1; i < len(ps); i++ {
		l = newMultiPolicty(l, ps[i])
	}

	return l
}

func newMultiPolicty(l, r EvictionPolicy) *multiPolicy {
	mp := multiPolicy{
		l:  l,
		r:  r,
		ch: make(chan string),
	}

	mp.wg.Add(1)
	mp.ctx, mp.cancel = context.WithCancel(context.Background())

	go mp.pull()

	return &mp
}

type multiPolicy struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan string

	l, r EvictionPolicy
}

func (mp *multiPolicy) pull() {
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

func (mp *multiPolicy) C() <-chan string {
	return mp.ch
}

func (mp *multiPolicy) Op(k string, op OpType) error {
	return multierror.Combine(mp.l.Op(k, op), mp.r.Op(k, op))
}

func (mp *multiPolicy) Close() error {
	mp.cancel()
	mp.wg.Wait()
	close(mp.ch)

	return multierror.Combine(mp.l.Close(), mp.r.Close())
}
