package record

import (
	"fmt"
	"io"
	"strconv"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 2048)
	},
}

type StringField struct {
	Key, Value string
}

func (s StringField) GetKey() string         { return s.Key }
func (s StringField) GetValue() string       { return s.Value }
func (s StringField) WriteKey(w io.Writer)   { io.WriteString(w, s.Key) }
func (s StringField) WriteValue(w io.Writer) { io.WriteString(w, s.Value) }

type Int64Field struct {
	Key   string
	Value int64
}

func (s Int64Field) GetKey() string { return s.Key }
func (s Int64Field) GetValue() string {
	return strconv.FormatInt(s.Value, 10)
}

func (s Int64Field) WriteKey(w io.Writer) { io.WriteString(w, s.Key) }
func (s Int64Field) WriteValue(w io.Writer) {
	buf := bufPool.Get().([]byte)

	w.Write(strconv.AppendInt(buf, s.Value, 10))

	buf = buf[:0]
	bufPool.Put(buf)
}

type BoolField struct {
	Key   string
	Value bool
}

func (s BoolField) GetKey() string       { return s.Key }
func (s BoolField) GetValue() string     { return strconv.FormatBool(s.Value) }
func (s BoolField) WriteKey(w io.Writer) { io.WriteString(w, s.Key) }
func (s BoolField) WriteValue(w io.Writer) {
	buf := bufPool.Get().([]byte)

	w.Write(strconv.AppendBool(buf, s.Value))

	buf = buf[:0]
	bufPool.Put(buf)
}

type Float64Field struct {
	Key   string
	Value float64
}

func (s Float64Field) GetKey() string { return s.Key }
func (s Float64Field) GetValue() string {
	return strconv.FormatFloat(s.Value, 'E', -1, 32)
}

func (s Float64Field) WriteKey(w io.Writer) { io.WriteString(w, s.Key) }
func (s Float64Field) WriteValue(w io.Writer) {
	buf := bufPool.Get().([]byte)

	w.Write(strconv.AppendFloat(buf, s.Value, 'E', -1, 32))

	buf = buf[:0]
	bufPool.Put(buf)
}

type StringerField struct {
	Key   string
	Value fmt.Stringer
}

func (s StringerField) GetKey() string       { return s.Key }
func (s StringerField) GetValue() string     { return s.Value.String() }
func (s StringerField) WriteKey(w io.Writer) { io.WriteString(w, s.Key) }
func (s StringerField) WriteValue(w io.Writer) {
	io.WriteString(w, s.Value.String())
}

type UnknownField struct {
	Key   string
	Value interface{}
	Fmt   string
}

func (s UnknownField) fmt() string {
	if s.Fmt != "" {
		return s.Fmt
	}

	return "%+v"
}

func (s UnknownField) GetKey() string { return s.Key }
func (s UnknownField) GetValue() string {
	return fmt.Sprintf(s.fmt(), s.Value)
}

func (s UnknownField) WriteKey(w io.Writer) { io.WriteString(w, s.Key) }
func (s UnknownField) WriteValue(w io.Writer) {
	fmt.Fprintf(w, s.fmt(), s.Value)
}
