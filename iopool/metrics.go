package iopool

import "github.com/upfluence/stats"

type metrics struct {
	get     stats.Counter
	put     stats.Counter
	discard stats.Counter

	getWait         stats.Counter
	getWaitDuration stats.Counter

	idleClosed stats.Counter

	idle     stats.Gauge
	checkout stats.Gauge
	size     stats.Gauge
}

func newMetrics(s stats.Scope) metrics {
	return metrics{
		get:             s.Counter("get_total"),
		put:             s.Counter("put_total"),
		discard:         s.Counter("discard_total"),
		getWait:         s.Counter("get_wait_total"),
		getWaitDuration: s.Counter("get_wait_millisecond_total"),
		idleClosed:      s.Counter("entity_idle_closed_total"),
		idle:            s.Gauge("idle"),
		checkout:        s.Gauge("checkout"),
		size:            s.Gauge("size"),
	}
}
