package multi

import (
	"github.com/upfluence/log/record"
	"github.com/upfluence/log/sink"
)

type Sink struct {
	ss []sink.Sink
}

func NewSink(ss ...sink.Sink) sink.Sink {
	if len(ss) == 1 {
		return ss[0]
	}

	return &Sink{ss: ss}
}

func (s *Sink) Log(r record.Record) error {
	for _, s := range s.ss {
		if err := s.Log(r); err != nil {
			return err
		}
	}

	return nil
}
