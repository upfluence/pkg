package log

import (
	"context"
	"fmt"
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

type Logger SugaredLogger

type logger struct {
	ctx record.Context

	sink      sink.Sink
	extractor ContextExtractor
	factory   record.RecordFactory

	defaultFieldThreshold record.Level
	defaultErrorThreshold record.Level
}

type LoggerOption func(*logger)

var (
	defaultContext   nullContext
	defaultExtractor noopExtractor

	defaultLogger = &logger{
		ctx:                   defaultContext,
		extractor:             defaultExtractor,
		factory:               record.NewFactory(),
		defaultFieldThreshold: record.Debug,
		defaultErrorThreshold: record.Debug,
	}
)

func WithSink(s sink.Sink) LoggerOption {
	return func(l *logger) { l.sink = s }
}

func WithDefaultFieldThreshold(lvl record.Level) LoggerOption {
	return func(l *logger) { l.defaultFieldThreshold = lvl }
}

func WithDefaultErrorThreshold(lvl record.Level) LoggerOption {
	return func(l *logger) { l.defaultErrorThreshold = lvl }
}

func WithContextExtractor(ce ContextExtractor) LoggerOption {
	return func(l *logger) {
		l.extractor = mergeContextExtractor(l.extractor, ce)
	}
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
		ctx:                   ctx,
		sink:                  l.sink,
		extractor:             l.extractor,
		factory:               l.factory,
		defaultFieldThreshold: l.defaultFieldThreshold,
		defaultErrorThreshold: l.defaultErrorThreshold,
	}
}

func (l *logger) WithField(f record.Field) SugaredLogger {
	if f == nil {
		return l
	}

	return l.WithFields(f)
}

func (l *logger) WithContext(ctx context.Context) SugaredLogger {
	if l.extractor == defaultExtractor || ctx == nil || ctx == context.Background() {
		return l
	}

	return l.dup(&withContext{Context: l.ctx, ctx: ctx, ce: l.extractor})
}

func (l *logger) WithFields(fs ...record.Field) SugaredLogger {
	if len(fs) == 0 {
		return l
	}

	return l.dup(
		&withFields{Context: l.ctx, fields: fs, threshold: l.defaultFieldThreshold},
	)
}

func (l *logger) WithError(err error) SugaredLogger {
	if err == nil {
		return l
	}

	return l.dup(
		&withErrors{
			Context:   l.ctx,
			errs:      []error{err},
			threshold: l.defaultErrorThreshold,
		},
	)
}

func (l *logger) Logf(lvl record.Level, fmt string, vs ...interface{}) {
	var r = l.factory.Build(l.ctx, lvl, fmt, vs...)

	l.sink.Log(r)
	l.factory.Free(r)
}

func (l *logger) Log(lvl record.Level, vs ...interface{}) {
	l.Logf(lvl, "", vs...)
}

func (l *logger) Debug(vs ...interface{})   { l.Debugf("", vs...) }
func (l *logger) Info(vs ...interface{})    { l.Infof("", vs...) }
func (l *logger) Notice(vs ...interface{})  { l.Noticef("", vs...) }
func (l *logger) Warning(vs ...interface{}) { l.Warningf("", vs...) }
func (l *logger) Error(vs ...interface{})   { l.Errorf("", vs...) }
func (l *logger) Fatal(vs ...interface{})   { l.Fatalf("", vs...) }

func (l *logger) Debugf(fmt string, vs ...interface{}) {
	l.Logf(record.Debug, fmt, vs...)
}

func (l *logger) Infof(fmt string, vs ...interface{}) {
	l.Logf(record.Info, fmt, vs...)
}

func (l *logger) Noticef(fmt string, vs ...interface{}) {
	l.Logf(record.Notice, fmt, vs...)
}

func (l *logger) Warningf(fmt string, vs ...interface{}) {
	l.Logf(record.Warning, fmt, vs...)
}

func (l *logger) Errorf(fmt string, vs ...interface{}) {
	l.Logf(record.Error, fmt, vs...)
}

func (l *logger) Fatalf(fmt string, vs ...interface{}) {
	l.Logf(record.Fatal, fmt, vs...)
	os.Exit(1)
}

func Field(k string, v interface{}) record.Field {
	switch vv := v.(type) {
	case string:
		return record.StringField{Key: k, Value: vv}
	case bool:
		return record.BoolField{Key: k, Value: vv}
	case int64:
		return record.Int64Field{Key: k, Value: vv}
	case int:
		return record.Int64Field{Key: k, Value: int64(vv)}
	case float64:
		return record.Float64Field{Key: k, Value: vv}
	case fmt.Stringer:
		return record.StringerField{Key: k, Value: vv}
	}

	return record.UnknownField{Key: k, Value: v}
}
