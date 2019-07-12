package stats

import (
	"math"
	"sync/atomic"
)

type atomicInt64 struct {
	int64
}

func (ai *atomicInt64) Inc()           { ai.Add(1) }
func (ai *atomicInt64) Add(v int64)    { atomic.AddInt64(&ai.int64, v) }
func (ai *atomicInt64) Get() int64     { return atomic.LoadInt64(&ai.int64) }
func (ai *atomicInt64) Update(v int64) { atomic.StoreInt64(&ai.int64, v) }

type atomicFloat64 struct {
	uint64
}

func (af *atomicFloat64) Get() float64 {
	return math.Float64frombits(atomic.LoadUint64(&af.uint64))
}

func (af *atomicFloat64) Update(v float64) {
	atomic.StoreUint64(&af.uint64, math.Float64bits(v))
}

func (af *atomicFloat64) Add(v float64) {
	for {
		old := atomic.LoadUint64(&af.uint64)

		if atomic.CompareAndSwapUint64(
			&af.uint64,
			old,
			math.Float64bits(math.Float64frombits(old)+v),
		) {
			return
		}
	}
}
