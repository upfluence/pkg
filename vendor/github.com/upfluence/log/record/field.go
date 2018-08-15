package record

import "fmt"

type StringField struct {
	Key, Value string
}

func (s *StringField) GetKey() string   { return s.Key }
func (s *StringField) GetValue() string { return s.Value }

type Int64Field struct {
	Key   string
	Value int64
}

func (s *Int64Field) GetKey() string   { return s.Key }
func (s *Int64Field) GetValue() string { return fmt.Sprintf("%d", s.Value) }
