package error_logger

import (
	"fmt"
	"os"

	"github.com/upfluence/errors/reporter"
	"github.com/upfluence/errors/reporter/inhibit"
	"github.com/upfluence/errors/reporter/sentry"

	"github.com/upfluence/pkg/peer"
	"github.com/upfluence/pkg/peer/version"
)

var (
	DefaultErrorLogger ErrorLogger
	DefaultReporter    Reporter
)

type reporterWrapper struct {
	r Reporter
}

func (rw *reporterWrapper) IgnoreErrors(fns ...func(error) bool) {
	eis := make([]inhibit.ErrorInhibitor, len(fns))

	for i, fn := range fns {
		eis[i] = inhibit.ErrorInhibitorFunc(fn)
	}

	rw.r.AddErrorInhibitors(eis...)
}

func (rw *reporterWrapper) Capture(err error, tags map[string]interface{}) error {
	rw.r.Report(err, reporter.ReportOptions{Tags: tags, Depth: 1})
	return nil
}

func (rw *reporterWrapper) Close() { _ = rw.r.Close() }

type Reporter interface {
	reporter.Reporter

	AddErrorInhibitors(...inhibit.ErrorInhibitor)
}

type ErrorLogger interface {
	IgnoreErrors(...func(error) bool)
	Capture(error, map[string]interface{}) error

	Close()
}

func init() {
	DefaultReporter = inhibit.NewReporter(buildReporter())
	DefaultErrorLogger = &reporterWrapper{r: DefaultReporter}

	if e := recover(); e != nil {
		if err, ok := e.(error); ok {
			DefaultReporter.Report(err, reporter.ReportOptions{})
			DefaultReporter.Close()
			panic(err.Error())
		}
	}
}

func buildReporter() reporter.Reporter {
	dsn := os.Getenv("SENTRY_DSN")

	if dsn == "" {
		return reporter.NopReporter
	}

	spo := sentryPeerOption{p: peer.FromEnv()}
	r, err := sentry.NewReporter(spo.apply)

	if err != nil {
		panic(err)
	}

	return r
}

type sentryPeerOption struct {
	p *peer.Peer
}

func (spo *sentryPeerOption) apply(opts *sentry.Options) {
	var tags = map[string]string{
		"semver_version": spo.p.Version.Semantic.String(),
		"unit_name":      spo.p.InstanceName,
		"environment":    spo.p.Environment,
	}

	if v := spo.p.Version.Git; v != (version.GitVersion{}) {
		tags["git_commit"] = v.Commit
		tags["git_branch"] = v.Branch
		tags["git_remote"] = v.Remote
	}

	for k, v := range tags {
		opts.Tags[k] = v
	}

	opts.SentryOptions.Release = fmt.Sprintf(
		"%s-%s",
		spo.p.ProjectName,
		spo.p.Version.String(),
	)
	opts.SentryOptions.Environment = spo.p.Environment
}

func Capture(err error, tags map[string]interface{}) error {
	return DefaultReporter.Report(
		err,
		reporter.ReportOptions{Tags: tags, Depth: 1},
	)
}

func Close() {
	DefaultReporter.Close()
}

func IgnoreError(err error) func(error) bool {
	return func(e error) bool { return e == err }
}
