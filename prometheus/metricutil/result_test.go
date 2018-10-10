package metricutil

import (
	stderrors "errors"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

type errMock struct{}

var errMockInstance = &errMock{}

func (m *errMock) Error() string { return "foo" }

func TestResultStatus(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   error
		out  string
	}{
		{name: "basic error", in: stderrors.New("error 1"), out: "error 1"},
		{name: "structured error", in: &errMock{}, out: "*metricutil.errMock"},

		{
			name: "with status error",
			in:   WrapStatus(stderrors.New("error 1"), "status"),
			out:  "status",
		},
		{
			name: "wrapped error",
			in:   pkgerrors.Wrap(stderrors.New("error 1"), "foo"),
			out:  "error 1",
		},

		{
			name: "wrapped error",
			in:   pkgerrors.Wrap(WrapStatus(stderrors.New("error 1"), "bar"), "foo"),
			out:  "bar",
		},

		{
			name: "static pkg/errors error",
			in:   pkgerrors.New("error 2"),
			out:  "error 2",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if out := ResultStatus(tt.in); tt.out != out {
				t.Errorf("ResultStatus(%v): %q [ want: %q ]", tt.in, out, tt.out)
			}
		})
	}
}
