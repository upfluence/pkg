package handler

import (
	"strings"

	"github.com/upfluence/base/monitoring"
	"github.com/upfluence/pkg/monitoring/metric"
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

func (m *MonitoringHandler) Collect(metrics []monitoring.MetricID) (
	monitoring.Metrics,
	error,
) {
	promises := make(map[monitoring.MetricID]chan []metric.Point)
	results := monitoring.Metrics{}

	for _, id := range metrics {
		if met, ok := m.metrics[id]; ok {
			result := make(chan []metric.Point)
			promises[id] = result

			go func() {
				result <- met.Collect()
			}()
		} else {
			return nil, &monitoring.UnknownMetric{Key: id}
		}
	}

	for id, promise := range promises {
		for _, point := range <-promise {
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
