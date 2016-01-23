package thrift

import (
	"github.com/upfluence/goutils/Godeps/_workspace/src/github.com/cyberdelia/statsd"
	"log"
	"os"
)

var (
	Metrics *Metric = NewMetric(os.Getenv("STATSD_URL"))
)

type Metric struct {
	client *statsd.Client
}

func NewMetric(statsdURL string) *Metric {
	var client *statsd.Client
	var err error

	if statsdURL != "" {
		client, err = statsd.Dial(statsdURL)

		if err != nil {
			log.Println(err.Error())

			return nil
		}
	}

	return &Metric{client}
}

func (m *Metric) Incr(metricName string) {
	if m.client != nil {
		m.client.Increment(metricName, 1, 1)
	}
}

func (m *Metric) Timing(metricName string, duration int64) {
	if m.client != nil {
		m.client.Timing(metricName, int(duration/1000000), 1)
	}
}
