package error_logger

import (
	"os"

	"github.com/upfluence/pkg/error_logger/noop"
	"github.com/upfluence/pkg/error_logger/sentry"
	"github.com/upfluence/pkg/peer"
)

var DefaultErrorLogger ErrorLogger

type ErrorLogger interface {
	IgnoreErrors(...func(error) bool)
	Capture(error, map[string]interface{}) error

	Close()
}

func IgnoreError(err error) func(error) bool {
	return func(e error) bool { return e == err }
}

func init() {
	if v := os.Getenv("SENTRY_DSN"); v != "" {
		l, err := sentry.NewErrorLogger(v, peer.FromEnv())

		if err != nil {
			DefaultErrorLogger = noop.NewErrorLogger()
		} else {
			DefaultErrorLogger = l
		}
	} else {
		DefaultErrorLogger = noop.NewErrorLogger()
	}

	if e := recover(); e != nil {
		if err, ok := e.(error); ok {
			DefaultErrorLogger.Capture(err, nil)
			DefaultErrorLogger.Close()
			panic(err.Error())
		}
	}
}

func Capture(err error, opts map[string]interface{}) error {
	return DefaultErrorLogger.Capture(err, opts)
}

func Close() {
	DefaultErrorLogger.Close()
}
