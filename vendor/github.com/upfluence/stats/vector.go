package stats

import "sync"

type Int64Value struct {
	Tags  map[string]string
	Value int64
}

type Int64VectorGetter interface {
	Labels() []string
	Get() []*Int64Value
}

type atomicInt64Vector struct {
	labels []string

	mu sync.RWMutex
	cs map[uint64]*atomicInt64

	marshaler labelMarshaler
}

func (v *atomicInt64Vector) Labels() []string { return v.labels }

func (v *atomicInt64Vector) buildTags(key uint64) map[string]string {
	var tags = make(map[string]string, len(v.labels))

	for i, val := range v.marshaler.unmarshal(key) {
		tags[v.labels[i]] = val
	}

	return tags
}

func (v *atomicInt64Vector) Get() []*Int64Value {
	var res = make([]*Int64Value, 0, len(v.cs))

	for k, c := range v.cs {
		res = append(res, &Int64Value{Tags: v.buildTags(k), Value: c.Get()})
	}

	return res
}

func (v *atomicInt64Vector) fetchValue(ls []string) *atomicInt64 {
	if len(ls) != len(v.labels) {
		panic("Not the correct number of labels")
	}

	k := v.marshaler.marshal(ls)

	v.mu.RLock()
	c, ok := v.cs[k]
	v.mu.RUnlock()

	if ok {
		return c
	}

	v.mu.Lock()
	c = &atomicInt64{}
	v.cs[k] = c
	v.mu.Unlock()

	return c
}
