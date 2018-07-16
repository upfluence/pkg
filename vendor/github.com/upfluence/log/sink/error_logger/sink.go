package error_logger

import (
	"errors"

	"github.com/upfluence/log/record"
	"github.com/upfluence/log/sink"
)

type ErrorLogger interface {
	Capture(error, map[string]interface{}) error
}

type Sink struct {
	eLogger ErrorLogger
}

func NewSink(el ErrorLogger) sink.Sink {
	return &Sink{eLogger: el}
}

func (s *Sink) Log(r record.Record) error {
	var (
		errs = r.Errs()
		tags = map[string]interface{}{}
	)

	if len(errs) == 0 {
		errs = []error{errors.New(r.Formatted())}
	}

	for _, f := range r.Fields() {
		tags[f.GetKey()] = f.GetValue()
	}

	for _, err := range errs {
		s.eLogger.Capture(err, tags)
	}

	return nil
}
