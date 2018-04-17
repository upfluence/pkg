package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/upfluence/pkg/pool"
	"github.com/upfluence/pkg/prometheus/metricutil"
)

const (
	namespace = "upfluence"
	subsystem = "pool"
)

var (
	callCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "call_total",
			Help:      "call count",
		},
		[]string{"name", "backend", "type", "status"},
	)

	callHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "call_duration_second",
			Help:      "call histogram",
		},
		[]string{"name", "backend", "type", "status"},
	)

	poolSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pool_size",
			Help:      "pool size gauges",
		},
		[]string{"name", "backend", "type"},
	)
)

func init() {
	prometheus.MustRegister(poolSize, callHistogram, callCount)
}

type PoolFactory struct {
	name    func(pool.Factory) string
	backend string
	next    pool.IntrospectablePoolFactory
}

func NewDefaultFactory(name string, next pool.IntrospectablePoolFactory) *PoolFactory {
	return &PoolFactory{
		name:    func(pool.Factory) string { return name },
		backend: fmt.Sprintf("%T", next),
		next:    next,
	}
}

func (f *PoolFactory) GetPool(factory pool.Factory) pool.Pool {
	return f.GetIntrospectablePool(factory)
}

func (f *PoolFactory) GetIntrospectablePool(factory pool.Factory) pool.IntrospectablePool {
	var p = &Pool{
		name:    f.name(factory),
		backend: f.backend,
	}

	p.next = f.next.GetIntrospectablePool(p.factory(factory))

	return p
}

type Pool struct {
	name, backend string
	next          pool.IntrospectablePool
}

func (p *Pool) GetStats() (int, int) { return p.next.GetStats() }

func (p *Pool) factory(f pool.Factory) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		var t0 = time.Now()

		e, err := f(ctx)

		p.reportAction("factory", err, t0)

		return e, err
	}
}

func (p *Pool) reportAction(t string, err error, t0 time.Time) {
	var labels = []string{
		p.name,
		p.backend,
		t,
		metricutil.ResultStatus(err),
	}

	callHistogram.WithLabelValues(labels...).Observe(time.Since(t0).Seconds())
	callCount.WithLabelValues(labels...).Inc()
}

func (p *Pool) updateGauges() {
	var in, out = p.next.GetStats()

	poolSize.WithLabelValues(p.name, p.backend, "checkin").Set(float64(in))
	poolSize.WithLabelValues(p.name, p.backend, "checkout").Set(float64(out))
}

func (p *Pool) Get(ctx context.Context) (interface{}, error) {
	var (
		t0     = time.Now()
		e, err = p.next.Get(ctx)
	)

	p.reportAction("get", err, t0)
	p.updateGauges()

	return e, err
}

func (p *Pool) Put(e interface{}) error {
	var (
		t0  = time.Now()
		err = p.next.Put(e)
	)

	p.reportAction("put", err, t0)
	p.updateGauges()

	return err
}

func (p *Pool) Discard(e interface{}) error {
	var (
		t0  = time.Now()
		err = p.next.Discard(e)
	)

	p.reportAction("discard", err, t0)
	p.updateGauges()

	return err
}
