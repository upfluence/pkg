package writer

import (
	"io"
	"os"

	"github.com/upfluence/log/record"
	"github.com/upfluence/log/sink"
)

type Formatter interface {
	Format(record.Record) string
}

type Sink struct {
	formatter Formatter
	writer    io.Writer
}

func NewStandardStdoutSink(cd int) sink.Sink {
	return NewStdoutSink(&formatter{calldepth: cd + 2})
}

func NewStdoutSink(f Formatter) sink.Sink {
	return &Sink{formatter: f, writer: os.Stdout}
}

func NewStandardSink(w io.Writer) sink.Sink {
	return NewSink(&formatter{calldepth: 3}, w)
}

func NewSink(f Formatter, w io.Writer) sink.Sink {
	return &Sink{formatter: f, writer: w}
}

func (s *Sink) Log(r record.Record) error {
	var _, err = s.writer.Write([]byte(s.formatter.Format(r) + "\n"))

	return err
}
