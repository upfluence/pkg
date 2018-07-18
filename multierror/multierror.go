package multierror

import (
	"fmt"
	"strings"
)

func Wrap(errs []error) error { return wrap(errs) }

func Combine(errs ...error) error { return wrap(errs) }

type MultiError interface {
	Errors() []error
}

type multiError []error

func wrap(errs []error) error {
	var out []error

	for _, err := range errs {
		if err == nil {
			continue
		}

		if merr, ok := err.(MultiError); ok {
			out = append(out, merr.Errors()...)
		} else {
			out = append(out, err)
		}
	}

	switch len(out) {
	case 0:
		return nil
	case 1:
		return out[0]
	default:
		return multiError(out)
	}
}

func (errs multiError) Errors() []error {
	return errs
}

func (errs multiError) Error() string {
	parts := make([]string, len(errs))

	for i, err := range errs {
		parts[i] = err.Error()
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}
