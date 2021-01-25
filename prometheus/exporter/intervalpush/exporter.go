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
	if interval == 0 {
		interval = defaultInterval
	}

	return &exporter{
		Monitor: closer.NewMonitor(closer.WithClosingPolicy(closer.Wait)),
		t:       time.NewTicker(interval),
		p: push.New(
			url,
			os.Getenv("APP_NAME"),
		).Gatherer(
			prometheus.DefaultGatherer,
		).Grouping("instance", os.Getenv("UNIT_NAME")),
	}
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
