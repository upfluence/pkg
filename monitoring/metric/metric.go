package metric

import "context"

type Point struct {
	// can be empty
	Suffix string
	Value  float64
}

type Metric interface {
	Collect(context.Context) []Point
}
