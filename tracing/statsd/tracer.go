package statsd

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	statsdClient "github.com/upfluence/goutils/Godeps/_workspace/src/github.com/cyberdelia/statsd"
)

const (
	defaultRateTime      = 0.1
	defaultRateIncrement = 1.0
)

type Tracer struct {
	client        *statsdClient.Client
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

func statsdfyName(name string) string {
	name = regexp.MustCompile("(-|@)").ReplaceAllString(name, "_")
	name = regexp.MustCompile("[^a-zA-Z0-9_\\.]").ReplaceAllString(name, "")
	return name
}

func (t *Tracer) Timing(name string, duration time.Duration) error {
	return t.client.Timing(
		statsdfyName(fmt.Sprintf("%s.%s", t.namespace, name)),
		int(duration/time.Millisecond),
		t.rateTime,
	)
}

func (t *Tracer) Trace(name string, fn func(), wg *sync.WaitGroup) error {
	if wg == nil {
		return t.client.Time(
			statsdfyName(fmt.Sprintf("%s.%s", t.namespace, name)),
			t.rateTime,
			fn,
		)
	} else {
		wg.Add(1)

		go func() {
			defer wg.Done()

			t.client.Time(
				statsdfyName(fmt.Sprintf("%s.%s", t.namespace, name)),
				t.rateTime,
				fn,
			)
		}()

		return nil
	}
}

func (t *Tracer) Count(bucket string, value int) error {
	return t.client.Increment(
		statsdfyName(fmt.Sprintf("%s.%s", t.namespace, bucket)),
		value,
		t.rateIncrement,
	)
}
