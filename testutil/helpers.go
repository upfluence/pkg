package testutil

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type ErrorAssertion func(testing.TB, error)

func NoError(msgAndArgs ...interface{}) ErrorAssertion {
	return func(t testing.TB, err error) {
		assert.Nil(t, err, msgAndArgs...)
	}
}

func ErrorEqual(err error, msgAndArgs ...interface{}) ErrorAssertion {
	return func(t testing.TB, gerr error) {
		assert.Equal(t, err, gerr, msgAndArgs...)
	}
}

func ErrorCause(err error, msgAndArgs ...interface{}) ErrorAssertion {
	return func(t testing.TB, gerr error) {
		assert.Equal(t, err, errors.Cause(gerr), msgAndArgs...)
	}
}

func FetchEnvVariable(t testing.TB, ev string) string {
	v := os.Getenv(ev)

	if v == "" {
		t.Skipf("%q env var is not defined, skipping test", ev)
	}

	return v
}
