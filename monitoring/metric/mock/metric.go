package mock

import (
	"github.com/upfluence/goutils/monitoring/metric"
)

type mockMetric struct {
	result float64
}

func NewMockMetric(result float64) metric.Metric {
	return &mockMetric{result}
}

func (m *mockMetric) Collect() <-chan float64 {
	out := make(chan float64)
	go func(out chan<- float64) {
		out <- m.result
	}(out)
	return out
}
