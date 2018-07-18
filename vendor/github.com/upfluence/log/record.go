package log

import "github.com/upfluence/log/record"

var SkipFrame = skipFrameField{}

type skipFrameField struct{}

func (skipFrameField) GetKey() string   { return "" }
func (skipFrameField) GetValue() string { return "" }
func (skipFrameField) SkipFrame() bool  { return true }

type withFields struct {
	record.Context

	fields []record.Field
}

func (wf *withFields) Fields() []record.Field {
	return append(wf.Context.Fields(), wf.fields...)
}

type withErrors struct {
	record.Context

	errs []error
}

func (we *withErrors) Errs() []error {
	return append(we.Context.Errs(), we.errs...)
}

type nullContext struct{}

func (*nullContext) Fields() []record.Field { return nil }
func (*nullContext) Errs() []error          { return nil }
