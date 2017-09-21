package intervalexporter

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/upfluence/pkg/log"
)

const defaultInterval = 15 * time.Second

type exporter struct {
	t                          *time.Ticker
	appName, pushURL, unitName string

	gatherer prometheus.Gatherer
}

func NewExporter(url string, interval time.Duration) *exporter {
	if interval == 0 {
		interval = defaultInterval
	}

	return &exporter{
		t:        time.NewTicker(interval),
		pushURL:  url,
		appName:  os.Getenv("APP_NAME"),
		unitName: os.Getenv("UNIT_NAME"),
		gatherer: prometheus.DefaultGatherer,
	}
}

func (e *exporter) Export(exitChan <-chan bool) {
	for {
		select {
		case <-exitChan:
			return
		case <-e.t.C:
			if err := push.FromGatherer(
				e.appName,
				map[string]string{"instance": e.unitName},
				e.pushURL,
				e.gatherer,
			); err != nil {
				log.Noticef("Push to gatherer failed: %s", err.Error())
			}
		}
	}
}
