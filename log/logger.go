package log

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/upfluence/log"
	"github.com/upfluence/log/record"
	elsink "github.com/upfluence/log/sink/error_logger"
	"github.com/upfluence/log/sink/leveled"
	"github.com/upfluence/log/sink/multi"
	"github.com/upfluence/log/sink/writer"

	"github.com/upfluence/pkg/v2/error_logger"
)

const localPkg = "github.com/upfluence/pkg/v2/log"

var (
	loggerMu   sync.Mutex
	loggerOpts = []log.LoggerOption{
		log.WithSink(
			multi.NewSink(
				leveled.NewSink(
					FetchLevel(),
					writer.NewStdoutSink(writer.NewDefaultFormatter(localPkg)),
				),
				leveled.NewSink(
					record.Error,
					elsink.WrapReporterWithBlacklist(
						error_logger.DefaultReporter,
						localPkg,
					),
				),
			),
		),
	}

	Logger = log.NewLogger(loggerOpts...)
)

func RegisterContextExtractor(ce log.ContextExtractor) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	loggerOpts = append(loggerOpts, log.WithContextExtractor(ce))
	Logger = log.NewLogger(loggerOpts...)
}

func FetchLevel() record.Level {
	switch strings.ToUpper(os.Getenv("LOGGER_LEVEL")) {
	case "DEBUG":
		return record.Debug
	case "INFO":
		return record.Info
	case "NOTICE":
		return record.Notice
	case "WARNING":
		return record.Warning
	case "ERROR":
		return record.Error
	case "FATAL":
		return record.Fatal
	}

	return record.Notice
}

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

var Field = log.Field
