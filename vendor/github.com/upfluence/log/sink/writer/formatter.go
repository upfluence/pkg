package writer

import (
	"io"

	"github.com/upfluence/log/internal/stacktrace"
	"github.com/upfluence/log/record"
)

const dateFmt = "020106 15:04:05"

var (
	levelPrettifier = map[record.Level][]byte{
		record.Debug:   []byte("D"),
		record.Info:    []byte("I"),
		record.Notice:  []byte("N"),
		record.Warning: []byte("W"),
		record.Error:   []byte("E"),
		record.Fatal:   []byte("F"),
	}

	defaultBlacklist = []string{"github.com/upfluence/log"}

	openBracket  = []byte("[")
	closeBracket = []byte("]")
	semiColon    = []byte(": ")
	errorKey     = []byte("[error: ")
	space        = []byte(" ")
)

func NewFastFormatter() Formatter {
	return &formatter{skipStacktrace: true}
}

func NewDefaultFormatter(blacklist ...string) Formatter {
	return &formatter{blacklist: append(defaultBlacklist, blacklist...)}
}

type formatter struct {
	blacklist      []string
	skipStacktrace bool
}

func newDefaultFormatter() *formatter {
	return &formatter{blacklist: defaultBlacklist}
}

func (f *formatter) formatFields(w io.Writer, fs []record.Field) {
	if len(fs) == 0 {
		return
	}

	for _, f := range fs {
		w.Write(openBracket)
		io.WriteString(w, f.GetKey())
		w.Write(semiColon)
		io.WriteString(w, f.GetValue())
		w.Write(closeBracket)
	}

	w.Write(space)
}

func (f *formatter) formatErrs(w io.Writer, errs []error) {
	if len(errs) == 0 {
		return
	}

	w.Write(space)

	for _, err := range errs {
		w.Write(errorKey)
		io.WriteString(w, err.Error())
		w.Write(closeBracket)
	}
}

func (f *formatter) Format(w io.Writer, r record.Record) error {
	w.Write(openBracket)
	w.Write(levelPrettifier[r.Level()])
	w.Write(space)
	w.Write(r.Time().AppendFormat(make([]byte, 0, len(dateFmt)), dateFmt))
	if !f.skipStacktrace {
		w.Write(space)
		stacktrace.WriteCaller(w, f.blacklist)
	}
	w.Write(closeBracket)
	w.Write(space)
	f.formatFields(w, r.Fields())
	r.WriteFormatted(w)
	f.formatErrs(w, r.Errs())

	return nil
}
