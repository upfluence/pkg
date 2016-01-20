package metric

type Metric interface {
	Collect() <-chan float64
}
