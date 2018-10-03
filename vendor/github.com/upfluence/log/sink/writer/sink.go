package writer

import (
	"io"
	"os"
	"sync"

	"github.com/upfluence/log/record"
	"github.com/upfluence/log/sink"
)

var newLine = []byte("\n")

type Formatter interface {
	Format(io.Writer, record.Record) error
}

type Sink struct {
	formatter Formatter
	writer    io.Writer
	mu        sync.Mutex
}

func NewStandardStdoutSink() sink.Sink {
	return NewStdoutSink(newDefaultFormatter())
}

func NewStdoutSink(f Formatter) sink.Sink {
	return &Sink{formatter: f, writer: os.Stdout}
}

func NewStandardSink(w io.Writer) sink.Sink {
	return NewSink(newDefaultFormatter(), w)
}

func NewSink(f Formatter, w io.Writer) sink.Sink {
	return &Sink{formatter: f, writer: w}
}

func (s *Sink) Log(r record.Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.formatter.Format(s.writer, r); err != nil {
		return err
	}

	_, err := s.writer.Write(newLine)

	return err
}
