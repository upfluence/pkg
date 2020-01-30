package log

import (
	"context"

	"github.com/upfluence/log/record"
)

type withFields struct {
	record.Context

	fields    []record.Field
	threshold record.Level
}

func (wf *withFields) Fields(lvl record.Level) []record.Field {
	if lvl < wf.threshold {
		return wf.Context.Fields(lvl)
	}

	return append(wf.Context.Fields(lvl), wf.fields...)
}

type withContext struct {
	record.Context

	ctx context.Context
	ce  ContextExtractor
}

func (wc *withContext) Fields(lvl record.Level) []record.Field {
	return append(wc.Context.Fields(lvl), wc.ce.Extract(wc.ctx, lvl)...)
}

type withErrors struct {
	record.Context

	errs      []error
	threshold record.Level
}

func (we *withErrors) Errs(lvl record.Level) []error {
	if lvl < we.threshold {
		return we.Context.Errs(lvl)
	}

	return append(we.Context.Errs(lvl), we.errs...)
}

type nullContext struct{}

func (nullContext) Fields(record.Level) []record.Field { return nil }
func (nullContext) Errs(record.Level) []error          { return nil }
