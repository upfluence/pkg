package record

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type Level uint

const (
	Debug Level = iota
	Info
	Notice
	Warning
	Error
	Fatal
)

type Field interface {
	GetKey() string
	GetValue() string
}

type Context interface {
	Fields() []Field
	Errs() []error
}

type Record interface {
	Context

	ID() uint64

	Time() time.Time
	Level() Level

	WriteFormatted(io.Writer)
	Args() []interface{}
}

type RecordFactory interface {
	Build(Context, Level, string, ...interface{}) Record
	Free(Record)
}

func NewFactory() RecordFactory {
	return &recordFactory{
		Pool: &sync.Pool{New: func() interface{} { return &record{} }},
	}
}

type recordFactory struct {
	*sync.Pool
}

var idCtr uint64

func (f *recordFactory) Free(r Record) {
	if v, ok := r.(*record); ok {
		f.Pool.Put(v)
	}
}

func (f *recordFactory) Build(ctx Context, l Level, fmt string, vs ...interface{}) Record {
	var r = f.Pool.Get().(*record)

	r.Context = ctx
	r.id = atomic.AddUint64(&idCtr, 1)
	r.t = time.Now()
	r.l = l
	r.fmt = fmt
	r.vs = vs

	return r
}

type record struct {
	Context

	id uint64
	t  time.Time
	l  Level

	fmt string
	vs  []interface{}
}

func (r *record) ID() uint64          { return r.id }
func (r *record) Time() time.Time     { return r.t }
func (r *record) Level() Level        { return r.l }
func (r *record) Args() []interface{} { return r.vs }

var space = []byte{' '}

func (r *record) WriteFormatted(w io.Writer) {
	if r.fmt == "" {
		for i, v := range r.vs {
			switch vv := v.(type) {
			case string:
				io.WriteString(w, vv)
			case []byte:
				w.Write(vv)
			default:
				fmt.Fprint(w, v)
			}

			if i != len(r.vs)-1 {
				w.Write(space)
			}
		}
		return
	}

	fmt.Fprintf(w, r.fmt, r.vs...)
}
