package testutil

import (
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
