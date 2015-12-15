package tracing

import "sync"

type Tracer interface {
	// if the third parameter is nil the closure will be executed synchronous
	// otherwise asynchronous
	Trace(string, func(), *sync.WaitGroup) error

	Count(bucket string, value int) error
}
