package mock

import (
	"context"

	"github.com/upfluence/pkg/monitoring/metric"
)

type mockMetric struct {
	result   float64
	suffixes []string
}

func NewMockMetric(result float64, suffixes []string) metric.Metric {
	return &mockMetric{result, suffixes}
}

func (m *mockMetric) Collect(_ context.Context) []metric.Point {
	r := []metric.Point{}

	for _, suffix := range m.suffixes {
		r = append(r, metric.Point{Suffix: suffix, Value: m.result})
	}

	return r
}
