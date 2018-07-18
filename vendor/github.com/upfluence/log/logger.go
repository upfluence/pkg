package log

import (
	"context"
	"os"

	"github.com/upfluence/log/record"
	"github.com/upfluence/log/sink"
)

type BasicLogger interface {
	Logf(record.Level, string, ...interface{})
	Log(record.Level, ...interface{})
}

type LeveledLogger interface {
	BasicLogger

	Debug(...interface{})
	Debugf(string, ...interface{})

	Info(...interface{})
	Infof(string, ...interface{})

	Notice(...interface{})
	Noticef(string, ...interface{})

	Warning(...interface{})
	Warningf(string, ...interface{})

	Error(...interface{})
	Errorf(string, ...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

type SugaredLogger interface {
	LeveledLogger

	WithField(record.Field) SugaredLogger
	WithFields(...record.Field) SugaredLogger
	WithContext(context.Context) SugaredLogger
	WithError(error) SugaredLogger
}

type ContextExtractor interface {
	Extract(context.Context) []record.Field
}

type noopExtractor struct{}

func (noopExtractor) Extract(context.Context) []record.Field { return nil }

type Logger SugaredLogger

type logger struct {
	ctx       record.Context
	sink      sink.Sink
	extractor ContextExtractor
	factory   record.RecordFactory
}

type LoggerOption func(*logger)

var defaultLogger = &logger{
	ctx:       &nullContext{},
	extractor: &noopExtractor{},
	factory:   record.NewFactory(),
}

func WithSink(s sink.Sink) LoggerOption {
	return func(l *logger) { l.sink = s }
}

func NewLogger(opts ...LoggerOption) Logger {
	var l = *defaultLogger

	for _, opt := range opts {
		opt(&l)
	}

	return &l
}

func (l *logger) dup(ctx record.Context) *logger {
	return &logger{
		ctx:       ctx,
		sink:      l.sink,
		extractor: l.extractor,
		factory:   l.factory,
	}
}

func (l *logger) WithField(f record.Field) SugaredLogger {
	return l.WithFields(f)
}

func (l *logger) WithContext(ctx context.Context) SugaredLogger {
	return l.WithFields(l.extractor.Extract(ctx)...)
}

func (l *logger) WithFields(fs ...record.Field) SugaredLogger {
	return l.dup(&withFields{Context: l.ctx, fields: fs})
}

func (l *logger) WithError(err error) SugaredLogger {
	return l.dup(&withErrors{Context: l.ctx, errs: []error{err}})
}

func (l *logger) Logf(lvl record.Level, fmt string, vs ...interface{}) {
	l.sink.Log(l.factory.Build(l.ctx, lvl, fmt, vs...))
}

func (l *logger) Log(lvl record.Level, vs ...interface{}) {
	l.WithField(SkipFrame).Log(lvl, "", vs)
}

func (l *logger) Debug(vs ...interface{})   { l.WithField(SkipFrame).Debugf("", vs...) }
func (l *logger) Info(vs ...interface{})    { l.WithField(SkipFrame).Infof("", vs...) }
func (l *logger) Notice(vs ...interface{})  { l.WithField(SkipFrame).Noticef("", vs...) }
func (l *logger) Warning(vs ...interface{}) { l.WithField(SkipFrame).Warningf("", vs...) }
func (l *logger) Error(vs ...interface{})   { l.WithField(SkipFrame).Errorf("", vs...) }
func (l *logger) Fatal(vs ...interface{})   { l.WithField(SkipFrame).Fatalf("", vs...) }

func (l *logger) Debugf(fmt string, vs ...interface{}) {
	l.WithField(SkipFrame).Logf(record.Debug, fmt, vs...)
}

func (l *logger) Infof(fmt string, vs ...interface{}) {
	l.WithField(SkipFrame).Logf(record.Info, fmt, vs...)
}

func (l *logger) Noticef(fmt string, vs ...interface{}) {
	l.WithField(SkipFrame).Logf(record.Notice, fmt, vs...)
}

func (l *logger) Warningf(fmt string, vs ...interface{}) {
	l.WithField(SkipFrame).Logf(record.Warning, fmt, vs...)
}

func (l *logger) Errorf(fmt string, vs ...interface{}) {
	l.WithField(SkipFrame).Logf(record.Error, fmt, vs...)
}

func (l *logger) Fatalf(fmt string, vs ...interface{}) {
	l.WithField(SkipFrame).Logf(record.Fatal, fmt, vs...)
	os.Exit(1)
}
