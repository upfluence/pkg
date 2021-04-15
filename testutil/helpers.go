package testutil

import (
	"os"
	"testing"

	"github.com/upfluence/errors/errtest"
)

type ErrorAssertion func(testing.TB, error)

func NoError(msgAndArgs ...interface{}) ErrorAssertion {
	return errtest.NoError(msgAndArgs...).Assert
}

func ErrorEqual(err error, msgAndArgs ...interface{}) ErrorAssertion {
	return errtest.ErrorEqual(err, msgAndArgs...).Assert
}

func ErrorCause(err error, msgAndArgs ...interface{}) ErrorAssertion {
	return errtest.ErrorCause(err, msgAndArgs...).Assert
}

func FetchEnvVariable(t testing.TB, ev string) string {
	v := os.Getenv(ev)

	if v == "" {
		t.Skipf("%q env var is not defined, skipping test", ev)
	}

	return v
}
