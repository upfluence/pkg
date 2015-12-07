package statsd

import (
	"fmt"
	"sync"

	statsdClient "github.com/cyberdelia/statsd"
)

const defaultRate = 0.1

type Tracer struct {
	*statsdClient.Client
	namespace string
	rate      float64
}

func NewTracer(statsdUrl, namespace string) (*Tracer, error) {
	c, err := statsdClient.Dial(statsdUrl)

	if err != nil {
		return nil, err
	}

	return &Tracer{c, namespace, defaultRate}, nil
}

func (t *Tracer) Trace(name string, fn func(), wg *sync.WaitGroup) error {
	if wg == nil {
		return t.Time(fmt.Sprintf("%s.%s", t.namespace, name), t.rate, fn)
	} else {
		wg.Add(1)

		go func() {
			defer wg.Done()

			t.Time(fmt.Sprintf("%s.%s", t.namespace, name), t.rate, fn)
		}()

		return nil
	}
}
