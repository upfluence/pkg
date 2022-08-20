package iopool

import (
	"time"

	"github.com/upfluence/pkg/cache/policy"
	ptime "github.com/upfluence/pkg/cache/policy/time"
	"github.com/upfluence/stats"
)

var defaultOptions = &options{
	size:     10,
	idleSize: 5,
	sc:       stats.RootScope(stats.NewStaticCollector()),
	scfn:     func(sc stats.Scope) stats.Scope { return sc },
}

type options struct {
	size     int
	idleSize int

	eps []policy.EvictionPolicy[uint64]

	sc   stats.Scope
	scfn func(stats.Scope) stats.Scope
}

func newOptions(opts ...Option) *options {
	var o = *defaultOptions

	for _, opt := range opts {
		opt(&o)
	}

	return &o
}

func (o *options) scope() stats.Scope { return o.scfn(o.sc) }

func (o *options) evictionPolicy() policy.EvictionPolicy[uint64] {
	return policy.CombinePolicies[uint64](o.eps...)
}

type Option func(*options)

func WithIdleTimeout(d time.Duration) Option {
	return func(o *options) {
		o.eps = append(o.eps, ptime.NewIdlePolicy[uint64](d))
	}
}

func WithScope(s stats.Scope) Option {
	return func(o *options) { o.sc = s }
}

func WithStandardizedMetrics(n string) Option {
	return func(o *options) {
		o.scfn = func(sc stats.Scope) stats.Scope {
			return sc.RootScope().Scope(
				"upfluence_iopool_pool",
				map[string]string{"pool": n},
			)
		}
	}
}

func WithMaxIdle(s int) Option {
	return func(o *options) {
		o.idleSize = s

		if s > o.size {
			o.size = s
		}
	}
}

func WithSize(s int) Option {
	return func(o *options) {
		o.size = s

		if o.idleSize > s {
			o.idleSize = s
		}
	}
}
