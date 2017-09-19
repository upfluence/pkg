package intervalexporter

import (
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/upfluence/pkg/log"
)

const (
	defaultInterval = 15 * time.Second
	defaultURL      = "http://localhost:9000"
)

type exporter struct {
	t                          *time.Ticker
	appName, pushURL, unitName string

	gatherer prometheus.Gatherer
}

func NewExporter() *exporter {
	return &exporter{
		t:        time.NewTicker(defaultInterval),
		pushURL:  defaultURL,
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
