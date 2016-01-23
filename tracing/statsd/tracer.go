package statsd

import (
	"fmt"
	"sync"

	statsdClient "github.com/upfluence/goutils/Godeps/_workspace/src/github.com/cyberdelia/statsd"
)

const (
	defaultRateTime      = 0.1
	defaultRateIncrement = 1.0
)

type Tracer struct {
	*statsdClient.Client
	namespace     string
	rateTime      float64
	rateIncrement float64
}

func NewTracer(statsdUrl, namespace string) (*Tracer, error) {
	c, err := statsdClient.Dial(statsdUrl)

	if err != nil {
		return nil, err
	}

	return &Tracer{c, namespace, defaultRateTime, defaultRateIncrement}, nil
}

func (t *Tracer) Trace(name string, fn func(), wg *sync.WaitGroup) error {
	if wg == nil {
		return t.Time(fmt.Sprintf("%s.%s", t.namespace, name), t.rateTime, fn)
	} else {
		wg.Add(1)

		go func() {
			defer wg.Done()

			t.Time(fmt.Sprintf("%s.%s", t.namespace, name), t.rateTime, fn)
		}()

		return nil
	}
}

func (t *Tracer) Count(bucket string, value int) error {
	return t.Increment(
		fmt.Sprintf("%s.%s", t.namespace, bucket),
		value,
		t.rateIncrement,
	)
}