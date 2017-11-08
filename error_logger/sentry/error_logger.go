package sentry

import (
	"fmt"
	"os"

	"github.com/getsentry/raven-go"
	"github.com/upfluence/pkg/thrift/handler"
)

type ErrorLogger struct {
	client *raven.Client

	errors []string
}

func NewErrorLogger(dsn string) (*ErrorLogger, error) {
	cl, err := raven.NewClient(
		dsn,
		map[string]string{
			"semver_version": handler.Version,
			"git_commit":     handler.GitCommit,
			"git_branch":     handler.GitBranch,
			"git_remote":     handler.GitRemote,
			"unit_name":      os.Getenv("UNIT_NAME"),
		},
	)

	if err != nil {
		return nil, err
	}

	if handler.Version != "v0.0.0" {
		cl.SetRelease(
			fmt.Sprintf("%s-%s", os.Getenv("PROJECT_NAME"), handler.Version),
		)
	}

	if v := os.Getenv("ENV"); v != "" {
		cl.SetEnvironment(v)
	}

	return &ErrorLogger{client: cl}, nil
}

func (l *ErrorLogger) IgnoreErrors(errs ...error) {
	for _, err := range errs {
		l.errors = append(l.errors, fmt.Sprintf("%T", err))
	}
}

func (l *ErrorLogger) Capture(err error, opts map[string]interface{}) error {
	var (
		errType = fmt.Sprintf("%T", err)
		tags    = make(map[string]string)
	)

	for _, ignoredError := range l.errors {
		if ignoredError == errType {
			return nil
		}
	}

	for k, v := range opts {
		tags[k] = fmt.Sprintf("%+v", v)
	}

	l.client.CaptureError(err, tags)

	return nil
}

func (l *ErrorLogger) Close() {
	l.client.Wait()
	l.client.Close()
}
