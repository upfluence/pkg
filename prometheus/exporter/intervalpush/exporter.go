package intervalexporter

import (
	"context"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/upfluence/pkg/closer"
	"github.com/upfluence/pkg/log"
	exp "github.com/upfluence/pkg/prometheus/exporter"
)

const defaultInterval = 15 * time.Second

type exporter struct {
	*closer.Monitor

	t *time.Ticker
	p *push.Pusher
}

func NewExporter(url string, interval time.Duration) exp.Exporter {
	return NewExporterWithOptions(
		url,
		WithInterval(interval),
		WithGrouping("instance", os.Getenv("UNIT_NAME")),
	)
}

type Option func(exporter *exporter)

func WithInterval(interval time.Duration) Option {
	return func(e *exporter) {
		e.t.Reset(interval)
	}
}

func WithEnvironment(env string) Option {
	return WithGrouping("env", env)
}

func WithAppName(name string) Option {
	return WithGrouping("app", name)
}

func WithProject(project string) Option {
	return WithGrouping("project", project)
}

func WithGrouping(name, value string) Option {
	return func(e *exporter) {
		e.p = e.p.Grouping(name, value)
	}
}

func NewExporterWithOptions(url string, opts ...Option) exp.Exporter {
	var e = exporter{
		Monitor: closer.NewMonitor(closer.WithClosingPolicy(closer.Wait)),
		t:       time.NewTicker(defaultInterval),
		p: push.New(
			url,
			os.Getenv("APP_NAME"),
		).Gatherer(
			prometheus.DefaultGatherer,
		),
	}

	for _, opt := range opts {
		opt(&e)
	}

	return &e
}

func (e *exporter) Export(exitChan <-chan bool) {
	e.Run(func(ctx context.Context) {
		for {
			select {
			case <-exitChan:
				e.t.Stop()
				return
			case <-ctx.Done():
				e.t.Stop()
				if err := e.p.Push(); err != nil {
					log.Noticef("Push to gatherer failed: %s", err.Error())
				}
				return
			case <-e.t.C:
				if err := e.p.Push(); err != nil {
					log.Noticef("Push to gatherer failed: %s", err.Error())
				}
			}
		}
	})
}
