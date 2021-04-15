package multierror

import (
	"github.com/upfluence/errors"
	"github.com/upfluence/errors/multi"
)

func Wrap(errs []error) error { return errors.WrapErrors(errs) }

func Combine(errs ...error) error { return errors.WrapErrors(errs) }

type MultiError = multi.MultiError
