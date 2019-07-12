package stats

import "fmt"

type Instrument interface {
	Exec(func() error) error
}

type InstrumentOption func(*instrumentOptions)

var defaultOptions = instrumentOptions{
	formatter:    defaultFormatter,
	trackStarted: true,
	counterLabel: "status",
}

func DisableStartedCounter() InstrumentOption {
	return func(opts *instrumentOptions) {
		opts.trackStarted = false
	}
}

func WithFormatter(f ErrorFormatter) InstrumentOption {
	return func(opts *instrumentOptions) {
		opts.formatter = f
	}
}

func WithCounterLabel(s string) InstrumentOption {
	return func(opts *instrumentOptions) {
		opts.counterLabel = s
	}
}

func WithTimerOptions(tOpts ...TimerOption) InstrumentOption {
	return func(opts *instrumentOptions) {
		opts.tOpts = tOpts
	}
}

type instrumentOptions struct {
	formatter    ErrorFormatter
	tOpts        []TimerOption
	trackStarted bool
	counterLabel string
}

func NewInstrument(scope Scope, name string, iOpts ...InstrumentOption) Instrument {
	var (
		opts = defaultOptions

		startedCounter Counter = &noopCounter{}
	)

	for _, opt := range iOpts {
		opt(&opts)
	}

	if opts.trackStarted {
		startedCounter = scope.Counter(fmt.Sprintf("%s_started_total", name))
	}

	return &instrument{
		instrumentOptions: opts,
		timer: NewTimer(
			scope,
			fmt.Sprintf("%s_duration", name),
			opts.tOpts...,
		),
		started: startedCounter,
		finished: scope.CounterVector(
			fmt.Sprintf("%s_total", name),
			[]string{opts.counterLabel},
		),
	}
}

type ErrorFormatter func(error) string

type instrument struct {
	instrumentOptions

	finished CounterVector
	started  Counter
	timer    Timer
}

func (i *instrument) Exec(fn func() error) error {
	i.started.Inc()
	sw := i.timer.Start()

	err := fn()

	sw.Stop()
	i.finished.WithLabels(i.formatter(err)).Inc()

	return err
}

func defaultFormatter(err error) string {
	if err == nil {
		return "success"
	}

	return err.Error()
}
