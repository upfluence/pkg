package sentry

import (
	"fmt"
	"os"

	"github.com/getsentry/raven-go"

	"github.com/upfluence/pkg/peer"
)

type ErrorLogger struct {
	client *raven.Client

	filters []func(error) bool
}

func NewErrorLogger(dsn string, p *peer.Peer) (*ErrorLogger, error) {
	var tags = map[string]string{
		"semver_version": peer.SerializeVersion(p.Version),
		"unit_name":      p.InstanceName,
		"environment":    p.Environment,
	}

	if v := p.Version.GitVersion; v != nil {
		tags["git_commit"] = v.Commit
		tags["git_branch"] = v.Branch
		tags["git_remote"] = v.Remote
	}

	cl, err := raven.NewClient(dsn, tags)

	if err != nil {
		return nil, err
	}

	cl.SetRelease(
		fmt.Sprintf("%s-%s", p.ProjectName, peer.SerializeVersion(p.Version)),
	)

	if v := os.Getenv("ENV"); v != "" {
		cl.SetEnvironment(v)
	}

	return &ErrorLogger{client: cl}, nil
}

func (l *ErrorLogger) IgnoreErrors(filters ...func(error) bool) {
	l.filters = append(l.filters, filters...)
}

func (l *ErrorLogger) errorIgnored(err error) bool {
	for _, filter := range l.filters {
		if filter(err) {
			return true
		}
	}

	return false
}

func splitError(err error) []error {
	var errs, ok = err.(interface{ Errors() []error })

	if !ok {
		return []error{err}
	}

	return errs.Errors()
}

func (l *ErrorLogger) Capture(err error, opts map[string]interface{}) error {
	var tags = make(map[string]string)

	if l.errorIgnored(err) {
		return nil
	}

	for k, v := range opts {
		tags[k] = fmt.Sprintf("%+v", v)
	}

	for _, nerr := range splitError(err) {
		l.client.CaptureError(nerr, tags)
	}

	return nil
}

func (l *ErrorLogger) Close() {
	l.client.Wait()
	l.client.Close()
}
