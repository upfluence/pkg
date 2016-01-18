package collector

import (
	"fmt"
	"strings"

	"github.com/upfluence/base/monitoring"
	"github.com/upfluence/sensu-client-go/sensu/check"
	"github.com/upfluence/sensu-client-go/sensu/handler"
)

type MetricsCollector struct {
	metrics []string
	root    string
	client  *monitoring.MonitoringClient
}

func NewMetricsCollector(
	raw, root string,
	client *monitoring.MonitoringClient,
) *MetricsCollector {
	return &MetricsCollector{strings.Split(raw, ","), root, client}
}

func (m *MetricsCollector) Collect() (check.ExtensionCheckResult, error) {
	metric := handler.Metric{}
	results, err := m.client.Collect(m.metrics)

	if err != nil {
		return metric.Render(), err
	}

	for key, val := range results {
		metric.AddPoint(
			&handler.Point{
				fmt.Sprintf("%s.%s", m.root, key),
				val,
			},
		)
	}

	return metric.Render(), nil
}
