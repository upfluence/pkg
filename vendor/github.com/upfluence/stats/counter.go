package stats

type CounterVector interface {
	WithLabels(...string) Counter
}

type Counter interface {
	Inc()
	Add(int64)

	Get() int64
}

type counterVector struct {
	*atomicInt64Vector
}

func (cv counterVector) WithLabels(ls ...string) Counter {
	return cv.fetchValue(ls)
}

type partialCounterVector struct {
	cv CounterVector
	vs []string
}

func (pcv partialCounterVector) WithLabels(ls ...string) Counter {
	return pcv.cv.WithLabels(append(pcv.vs, ls...)...)
}

type reorderCounterVector struct {
	cv CounterVector
	labelOrderer
}

func (rcv reorderCounterVector) WithLabels(ls ...string) Counter {
	return rcv.cv.WithLabels(rcv.order(ls)...)
}

type noopCounter struct{}

func (noopCounter) Inc()       {}
func (noopCounter) Add(int64)  {}
func (noopCounter) Get() int64 { return 0 }

type noopCounterVector struct{}

func (noopCounterVector) WithLabels(...string) Counter { return &noopCounter{} }
