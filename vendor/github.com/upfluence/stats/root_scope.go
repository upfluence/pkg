package stats

import (
	"fmt"
	"sort"
	"sync"
)

type rootScope struct {
	c Collector
	sync.Mutex

	counters   map[string]*atomicInt64Vector
	gauges     map[string]*atomicInt64Vector
	histograms map[string]*histogramVector
}

func RootScope(c Collector) Scope {
	return scopeWrapper{
		&rootScope{
			c:          c,
			counters:   make(map[string]*atomicInt64Vector),
			gauges:     make(map[string]*atomicInt64Vector),
			histograms: make(map[string]*histogramVector),
		},
	}
}

func (rs *rootScope) assertMetricUniqueness(n string) {
	if _, ok := rs.counters[n]; ok {
		panic(fmt.Sprintf("counter with the same name already registered: %q", n))
	}

	if _, ok := rs.gauges[n]; ok {
		panic(fmt.Sprintf("gauge with the same name already registered: %q", n))
	}

	if _, ok := rs.histograms[n]; ok {
		panic(fmt.Sprintf("hisogram with the same name already registered: %q", n))
	}
}

type labelOrderer interface {
	order([]string) []string
}

type mappingLabelOdrderer struct {
	mapping map[int]int
}

func (mlo mappingLabelOdrderer) order(in []string) []string {
	if len(mlo.mapping) != len(in) {
		panic("wrong number of labels passed")
	}

	var out = make([]string, len(in))

	for i, v := range in {
		out[mlo.mapping[i]] = v
	}

	return out
}

func buildLabelOrderer(base, target []string) labelOrderer {
	if len(base) != len(target) {
		panic("Another metric is already created with different number of label")
	}

	var (
		baseCopy   = append(base[:0:0], base...)
		targetCopy = append(target[:0:0], target...)
	)

	sort.Sort(sort.StringSlice(baseCopy))
	sort.Sort(sort.StringSlice(targetCopy))

	for i, bv := range baseCopy {
		if targetCopy[i] != bv {
			panic("Another metric is already created with different label")
		}
	}

	var transferMap = make(map[int]int, len(base))

	for i, tv := range target {
		for j, bv := range base {
			if tv == bv {
				transferMap[i] = j
				break
			}
		}
	}

	return mappingLabelOdrderer{mapping: transferMap}
}

func (rs *rootScope) registerHistogram(n string, ls []string, opts ...HistogramOption) HistogramVector {
	rs.Lock()
	defer rs.Unlock()

	if h, ok := rs.histograms[n]; ok {
		// TODO: ensure the cutoffs are similare or panic

		return reorderHistogramVector{
			hv:           h,
			labelOrderer: buildLabelOrderer(h.labels, ls),
		}
	}

	rs.assertMetricUniqueness(n)

	v := &histogramVector{
		cutoffs:   defaultCutoffs,
		labels:    ls,
		hs:        map[uint64]*histogram{},
		marshaler: newDefaultMarshaler(),
	}

	for _, opt := range opts {
		opt(v)
	}

	rs.histograms[n] = v
	rs.c.RegisterHistogram(n, v)

	return v
}

func (rs *rootScope) registerGauge(n string, ls []string) GaugeVector {
	rs.Lock()
	defer rs.Unlock()

	if c, ok := rs.gauges[n]; ok {
		return reorderGaugeVector{
			gv:           gaugeVector{c},
			labelOrderer: buildLabelOrderer(c.labels, ls),
		}
	}

	rs.assertMetricUniqueness(n)

	v := &atomicInt64Vector{
		labels:    ls,
		cs:        make(map[uint64]*atomicInt64),
		marshaler: newDefaultMarshaler(),
	}

	rs.gauges[n] = v
	rs.c.RegisterGauge(n, v)

	return gaugeVector{v}
}

func (rs *rootScope) registerCounter(n string, ls []string) CounterVector {
	rs.Lock()
	defer rs.Unlock()

	if c, ok := rs.counters[n]; ok {
		return reorderCounterVector{
			cv:           counterVector{c},
			labelOrderer: buildLabelOrderer(c.labels, ls),
		}
	}

	rs.assertMetricUniqueness(n)

	v := &atomicInt64Vector{
		labels:    ls,
		cs:        make(map[uint64]*atomicInt64),
		marshaler: newDefaultMarshaler(),
	}

	rs.counters[n] = v
	rs.c.RegisterCounter(n, v)

	return counterVector{v}
}

func (*rootScope) namespace() string        { return "" }
func (*rootScope) tags() map[string]string  { return nil }
func (rs *rootScope) rootScope() *rootScope { return rs }
