package stats

import (
	"fmt"
	"sync"
	"time"
)

type StopWatch interface {
	Stop()
}

type Timer interface {
	Start() StopWatch
}

type timer struct {
	Histogram

	p sync.Pool
}

func (t *timer) Start() StopWatch {
	var sw = t.p.Get().(*stopWatch)

	sw.t0 = time.Now()

	return sw
}

type stopWatch struct {
	t0    time.Time
	timer *timer
}

type TimerOption func(*timerOptions)

type timerOptions struct {
	hOpts  []HistogramOption
	suffix string
}

func WithHistogramOptions(hOpts ...HistogramOption) TimerOption {
	return func(opts *timerOptions) {
		opts.hOpts = hOpts
	}
}

func WithTimerSuffix(s string) TimerOption {
	return func(opts *timerOptions) {
		opts.suffix = s
	}
}

var defaultTimerOptions = timerOptions{
	suffix: "_seconds",
}

func (sw *stopWatch) Stop() {
	sw.timer.Record(time.Since(sw.t0).Seconds())
	sw.timer.p.Put(sw)
}

func NewTimer(scope Scope, name string, tOpts ...TimerOption) Timer {
	var opts = defaultTimerOptions

	for _, opt := range tOpts {
		opt(&opts)
	}

	var t = &timer{
		Histogram: scope.Histogram(
			fmt.Sprintf("%s%s", name, opts.suffix),
			opts.hOpts...,
		),
	}

	t.p = sync.Pool{New: func() interface{} { return &stopWatch{timer: t} }}

	return t
}
