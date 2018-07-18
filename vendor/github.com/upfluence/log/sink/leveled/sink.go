package leveled

import (
	"github.com/upfluence/log/record"
	"github.com/upfluence/log/sink"
)

type Sink struct {
	l record.Level
	s sink.Sink
}

func NewSink(l record.Level, s sink.Sink) sink.Sink {
	return &Sink{l: l, s: s}
}

func (s *Sink) Log(r record.Record) error {
	if r.Level() < s.l {
		return nil
	}

	return s.s.Log(r)
}
