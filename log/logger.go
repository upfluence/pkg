package log

import (
	"context"

	"github.com/upfluence/log"
	"github.com/upfluence/log/record"
	elsink "github.com/upfluence/log/sink/error_logger"
	"github.com/upfluence/log/sink/multi"
	"github.com/upfluence/log/sink/writer"

	"github.com/upfluence/pkg/error_logger"
)

var (
	Logger = log.NewLogger(
		log.WithSink(
			multi.NewSink(
				writer.NewStandardStdoutSink(2),
				elsink.NewSink(error_logger.DefaultErrorLogger),
			),
		),
	)
)

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

func Warning(args ...interface{}) {
	Logger.Warning(args...)
}

func Warningf(format string, args ...interface{}) {
	Logger.Warningf(format, args...)
}

func Notice(args ...interface{}) {
	Logger.Notice(args...)
}

func Noticef(format string, args ...interface{}) {
	Logger.Noticef(format, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

func WithField(f record.Field) log.Logger {
	return Logger.WithField(f)
}

func WithFields(fs ...record.Field) log.Logger {
	return Logger.WithFields(fs...)
}

func WithContext(ctx context.Context) log.Logger {
	return Logger.WithContext(ctx)
}

func WithError(err error) log.Logger {
	return Logger.WithError(err)
}
