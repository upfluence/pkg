package noop

import "github.com/upfluence/goutils/error_logger"

type Logger struct{}

func NewErrorLogger() *Logger { return &Logger{} }

func (l *Logger) Capture(err error, opts *error_logger.Options) error { return nil }
func (l *Logger) Close()                                              {}
