package record

import (
	"fmt"
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

	Formatted() string
}

type RecordFactory interface {
	Build(Context, Level, string, ...interface{}) Record
}

func NewFactory() RecordFactory { return &recordFactory{} }

type recordFactory struct{}

var idCtr uint64

func (*recordFactory) Build(ctx Context, l Level, fmt string, vs ...interface{}) Record {
	return &record{
		Context: ctx,
		id:      atomic.AddUint64(&idCtr, 1),
		t:       time.Now(),
		l:       l,
		fmt:     fmt,
		vs:      vs,
	}
}

type record struct {
	Context

	id uint64
	t  time.Time
	l  Level

	fmt string
	vs  []interface{}
}

func (r *record) ID() uint64      { return r.id }
func (r *record) Time() time.Time { return r.t }
func (r *record) Level() Level    { return r.l }

func (r *record) Formatted() string {
	if r.fmt == "" {
		return fmt.Sprint(r.vs...)
	}

	return fmt.Sprintf(r.fmt, r.vs...)
}
