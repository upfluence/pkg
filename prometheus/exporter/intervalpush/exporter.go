package intervalexporter

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/upfluence/pkg/log"
	exp "github.com/upfluence/pkg/prometheus/exporter"
)

const defaultInterval = 15 * time.Second

type exporter struct {
	t *time.Ticker
	p *push.Pusher
}

func NewExporter(url string, interval time.Duration) exp.Exporter {
	if interval == 0 {
		interval = defaultInterval
	}

	return &exporter{
		t: time.NewTicker(interval),
		p: push.New(
			url,
			os.Getenv("APP_NAME"),
		).Gatherer(
			prometheus.DefaultGatherer,
		).Grouping("instance", os.Getenv("UNIT_NAME")),
	}
}

func (e *exporter) Export(exitChan <-chan bool) {
	for {
		select {
		case <-exitChan:
			return
		case <-e.t.C:
			if err := e.p.Push(); err != nil {
				log.Noticef("Push to gatherer failed: %s", err.Error())
			}
		}
	}
}
