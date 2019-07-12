package stats

import "io"

type Collector interface {
	io.Closer

	RegisterCounter(string, Int64VectorGetter)
	RegisterGauge(string, Int64VectorGetter)
	RegisterHistogram(string, HistogramVectorGetter)
}
