package multierror

import "github.com/upfluence/errors"

func Wrap(errs []error) error { return errors.Wrap(errs) }

func Combine(errs ...error) error { return errors.Wrap(errs) }

type MultiError interface {
	Errors() []error
}
