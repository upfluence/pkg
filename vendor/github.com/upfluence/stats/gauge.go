package stats

type GaugeVector interface {
	WithLabels(...string) Gauge
}

type Gauge interface {
	Update(int64)
	Get() int64
}

type gaugeVector struct {
	*atomicInt64Vector
}

func (gv gaugeVector) WithLabels(ls ...string) Gauge {
	return gv.fetchValue(ls)
}

type partialGaugeVector struct {
	gv GaugeVector
	vs []string
}

func (pgv partialGaugeVector) WithLabels(ls ...string) Gauge {
	return pgv.gv.WithLabels(append(pgv.vs, ls...)...)
}

type reorderGaugeVector struct {
	gv GaugeVector
	labelOrderer
}

func (rgv reorderGaugeVector) WithLabels(ls ...string) Gauge {
	return rgv.gv.WithLabels(rgv.order(ls)...)
}
