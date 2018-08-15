package error_logger

import (
	"bytes"
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
		for _, arg := range r.Args() {
			if err, ok := arg.(error); ok {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) == 0 {
		var buf bytes.Buffer
		r.WriteFormatted(&buf)

		errs = []error{errors.New(buf.String())}
	}

	for _, f := range r.Fields() {
		tags[f.GetKey()] = f.GetValue()
	}

	for _, err := range errs {
		s.eLogger.Capture(err, tags)
	}

	return nil
}
