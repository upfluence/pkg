package error_logger

import (
	"errors"
	"fmt"
	"os"

	"github.com/upfluence/errors/reporter"
	"github.com/upfluence/errors/reporter/inhibit"
	"github.com/upfluence/errors/reporter/sentry"

	"github.com/upfluence/pkg/v2/peer"
	"github.com/upfluence/pkg/v2/peer/version"
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

func (rw *reporterWrapper) WhitelistTag(fns ...func(string) bool) {
	rw.r.WhitelistTag(fns...)
}

func (rw *reporterWrapper) Close() { _ = rw.r.Close() }

type TagWhitelister interface {
	WhitelistTag(...func(string) bool)
}

type noopTagWhitelister struct{}

func (noopTagWhitelister) WhitelistTag(...func(string) bool) {}

type ErrorIgnorer interface {
	AddErrorInhibitors(...inhibit.ErrorInhibitor)
}

type Reporter interface {
	reporter.Reporter
	TagWhitelister
	ErrorIgnorer
}

type reporterImpl struct {
	reporter.Reporter
	TagWhitelister
	ErrorIgnorer
}

func expandReporter(r reporter.Reporter) Reporter {
	impl := reporterImpl{Reporter: r, TagWhitelister: noopTagWhitelister{}}

	if tw, ok := r.(TagWhitelister); ok {
		impl.TagWhitelister = tw
	}

	if ei, ok := r.(ErrorIgnorer); ok {
		impl.ErrorIgnorer = ei
	} else {
		eir := inhibit.NewReporter(r)

		impl.ErrorIgnorer = eir
		impl.Reporter = eir
	}

	return &impl
}

type ErrorLogger interface {
	IgnoreErrors(...func(error) bool)
	WhitelistTag(...func(string) bool)
	Capture(error, map[string]interface{}) error

	Close()
}

func init() {
	DefaultReporter = expandReporter(buildReporter())
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
	DefaultReporter.Report(err, reporter.ReportOptions{Tags: tags, Depth: 1})

	return nil
}

func Close() {
	DefaultReporter.Close()
}

func IgnoreError(err error) func(error) bool {
	return func(e error) bool { return errors.Is(e, err) }
}
