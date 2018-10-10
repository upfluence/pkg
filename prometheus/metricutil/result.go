package metricutil

import "fmt"

const successStatus = staticStatuser("success")

type statuser interface {
	Status() string
}

type causer interface {
	Cause() error
}

type staticStatuser string

func (s staticStatuser) Status() string { return string(s) }

type errorStatuser struct {
	err error
}

func (w *errorStatuser) Status() string {
	switch t := fmt.Sprintf("%T", w.err); t {
	case "*errors.errorString", "*errors.fundamental":
		return w.err.Error()
	default:
		return t
	}
}

type withStatus struct {
	error
	status string
}

func (w *withStatus) Cause() error   { return w.error }
func (w *withStatus) Status() string { return w.status }

func WrapStatus(err error, status string) error {
	return &withStatus{error: err, status: status}
}

func ResultStatus(err error) string {
	var res statuser = successStatus

	for err != nil {
		if status, ok := err.(statuser); ok {
			res = status
			break
		}

		cause, ok := err.(causer)

		if !ok {
			res = &errorStatuser{err}
			break
		}

		err = cause.Cause()
	}

	return res.Status()
}
