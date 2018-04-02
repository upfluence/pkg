package handler

import (
	"strings"

	"github.com/upfluence/base/monitoring"
	"github.com/upfluence/pkg/monitoring/metric"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type MonitoringHandler struct {
	prefix  string
	metrics map[monitoring.MetricID]metric.Metric
}

func NewMonitoringHandler(
	prefix string,
	metrics map[monitoring.MetricID]metric.Metric,
) *MonitoringHandler {
	return &MonitoringHandler{prefix, metrics}
}

func (m *MonitoringHandler) Collect(ctx thrift.Context, metrics []monitoring.MetricID) (
	monitoring.Metrics,
	error,
) {
	promises := make(map[monitoring.MetricID]chan metric.Point)
	results := monitoring.Metrics{}

	for _, id := range metrics {
		if met, ok := m.metrics[id]; ok {
			result := make(chan metric.Point)
			promises[id] = result

			go func() {
				for _, p := range met.Collect(ctx) {
					result <- p
				}

				close(result)
			}()
		} else {
			return nil, &monitoring.UnknownMetric{Key: id}
		}
	}

	for id, promise := range promises {
		for point := range promise {
			splittedName := []string{m.prefix, string(id)}
			if v := point.Suffix; v != "" {
				splittedName = append(splittedName, v)
			}

			metricName := monitoring.MetricID(strings.Join(splittedName, "."))
			results[metricName] = point.Value
		}
	}

	return results, nil
}
