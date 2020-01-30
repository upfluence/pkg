package stats

import (
	"strings"
)

type limitedScope interface {
	namespace() string
	tags() map[string]string

	rootScope() *rootScope
}

type HistogramOption func(*histogramVector)

type Scope interface {
	limitedScope

	Counter(string) Counter
	CounterVector(string, []string) CounterVector

	Gauge(string) Gauge
	GaugeVector(string, []string) GaugeVector

	Histogram(string, ...HistogramOption) Histogram
	HistogramVector(string, []string, ...HistogramOption) HistogramVector

	Scope(string, map[string]string) Scope
	RootScope() Scope
}

type scopeWrapper struct {
	limitedScope
}

func (sw scopeWrapper) buildLabelValues() ([]string, []string) {
	var (
		tags   = sw.tags()
		labels = make([]string, 0, len(tags))
		values = make([]string, 0, len(tags))
	)

	for l, v := range tags {
		labels = append(labels, l)
		values = append(values, v)
	}

	return labels, values
}

func (sw scopeWrapper) Histogram(name string, opts ...HistogramOption) Histogram {
	var (
		ls, vs = sw.buildLabelValues()

		hv = sw.rootScope().registerHistogram(
			joinStrings(sw.namespace(), name),
			ls,
			opts...,
		)
	)

	return hv.WithLabels(vs...)

}

func (sw scopeWrapper) HistogramVector(name string, labels []string, opts ...HistogramOption) HistogramVector {
	var (
		sls, vs = sw.buildLabelValues()

		hv = sw.rootScope().registerHistogram(
			joinStrings(sw.namespace(), name),
			append(sls, labels...),
			opts...,
		)
	)

	return partialHistogramVector{hv: hv, vs: vs}
}

func (sw scopeWrapper) Gauge(name string) Gauge {
	var (
		ls, vs = sw.buildLabelValues()

		gv = sw.rootScope().registerGauge(joinStrings(sw.namespace(), name), ls)
	)

	return gv.WithLabels(vs...)
}

func (sw scopeWrapper) GaugeVector(name string, labels []string) GaugeVector {
	var (
		sls, vs = sw.buildLabelValues()

		gv = sw.rootScope().registerGauge(
			joinStrings(sw.namespace(), name),
			append(sls, labels...),
		)
	)

	return partialGaugeVector{gv: gv, vs: vs}
}

func (sw scopeWrapper) Counter(name string) Counter {
	var (
		ls, vs = sw.buildLabelValues()

		cv = sw.rootScope().registerCounter(joinStrings(sw.namespace(), name), ls)
	)

	return cv.WithLabels(vs...)
}

func (sw scopeWrapper) CounterVector(name string, labels []string) CounterVector {
	var (
		pls, vs = sw.buildLabelValues()

		cv = sw.rootScope().registerCounter(
			joinStrings(sw.namespace(), name),
			append(pls, labels...),
		)
	)

	return partialCounterVector{cv: cv, vs: vs}
}

func (sw scopeWrapper) Scope(ns string, tags map[string]string) Scope {
	return scopeWrapper{
		limitedScope: &subScope{parent: sw.limitedScope, ns: ns, ts: tags},
	}
}

func (sw scopeWrapper) RootScope() Scope {
	return scopeWrapper{limitedScope: sw.rootScope()}
}

type subScope struct {
	parent limitedScope

	ns string
	ts map[string]string
}

func (ss *subScope) namespace() string {
	return joinStrings(ss.parent.namespace(), ss.ns)
}

func (ss *subScope) tags() map[string]string {
	return mergeStringMaps(ss.parent.tags(), ss.ts)
}

func (ss *subScope) rootScope() *rootScope { return ss.parent.rootScope() }

func joinStrings(ss ...string) string {
	var (
		res []string

		append = func(s string) {
			if s != "" {
				res = append(res, s)
			}
		}
	)

	for _, s := range ss {
		append(s)
	}

	return strings.Join(res, "_")
}

func mergeStringMaps(kvs ...map[string]string) map[string]string {
	var (
		res = make(map[string]string)

		merge = func(kv map[string]string) {
			for k, v := range kv {
				res[k] = v
			}
		}
	)

	for _, kv := range kvs {
		merge(kv)
	}

	return res
}
