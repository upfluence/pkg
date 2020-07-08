package setter

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/upfluence/cfg/internal/reflectutil"
)

var (
	durationType = reflect.TypeOf(time.Duration(0))
	timeType     = reflect.TypeOf(time.Time{})
	ValueType    = reflect.TypeOf((*Value)(nil)).Elem()

	presetParsers = map[reflect.Type]parser{
		durationType: durationParser{},
		timeType:     timeParser{},
	}

	DefaultFactory Factory = defaultFactory{}
)

type Value interface {
	Parse(string) error
}

type Factory interface {
	Build(reflect.StructField) Setter
}

type defaultFactory struct{}

func (dsf defaultFactory) buildBasicParser(t reflect.Type) (parser, bool) {
	var (
		k  = t.Kind()
		pt = t

		ptr bool
	)

	if k == reflect.Ptr {
		k = t.Elem().Kind()
		ptr = true
		t = t.Elem()
	} else {
		pt = reflect.PtrTo(t)
	}

	if pt.Implements(ValueType) {
		return &valueParser{t: t}, ptr
	}

	if p, ok := presetParsers[t]; ok {
		return p, ptr
	}

	switch k {
	case reflect.String:
		return stringParser{}, ptr
	case reflect.Int, reflect.Int64:
		return &intParser{transformer: intTransformers[k]}, ptr
	case reflect.Float64:
		return floatParser{}, ptr
	case reflect.Bool:
		return boolParser{}, ptr
	}

	return nil, false
}

func (dsf defaultFactory) buildParser(t reflect.Type) parser {
	k := t.Kind()

	switch k {
	case reflect.Slice:
		p, ptr := dsf.buildBasicParser(t.Elem())

		if p == nil {
			return nil
		}

		return &sliceParser{p: p, t: t, ptr: ptr}
	case reflect.Map:
		vp, vptr := dsf.buildBasicParser(t.Elem())

		if vp == nil {
			return nil
		}

		kp, kptr := dsf.buildBasicParser(t.Key())

		if kp == nil {
			return nil
		}

		return &mapParser{t: t, vp: vp, vptr: vptr, kp: kp, kptr: kptr}
	}

	p, _ := dsf.buildBasicParser(t)

	return p
}

func (df defaultFactory) Build(f reflect.StructField) Setter {
	if p := df.buildParser(reflectutil.IndirectedType(f.Type)); p != nil {
		return &parserSetter{field: f, parser: p}
	}

	return nil
}

type Setter interface {
	fmt.Stringer

	Set(string, interface{}) error
}

type ErrSetterNotImplemented struct {
	field reflect.StructField
}

func (e *ErrSetterNotImplemented) Error() string {
	return fmt.Sprintf("cfg: Setter not implemented for type %v", e.field.Type)
}

type boolParser struct{}

type ErrNotBoolValue struct {
	value string
}

func (e *ErrNotBoolValue) Error() string {
	return fmt.Sprintf("cfg: Can't parse %q in a bool value", e.value)
}

func (boolParser) String() string { return "bool" }

func (boolParser) parse(value string, ptr bool) (interface{}, error) {
	var v bool

	switch strings.TrimSpace(value) {
	case "t", "1", "true", "yes", "y":
		v = true
	case "f", "0", "false", "no", "n":
	default:
		return nil, &ErrNotBoolValue{value: value}
	}

	if ptr {
		return &v, nil
	}

	return v, nil
}

type parserSetter struct {
	field  reflect.StructField
	parser parser
}

func (s *parserSetter) String() string { return s.parser.String() }

func (s *parserSetter) Set(value string, target interface{}) error {
	var t = reflectutil.IndirectedValue(reflect.ValueOf(target)).FieldByName(s.field.Name)

	v, err := s.parser.parse(value, t.Type().Kind() == reflect.Ptr)

	if err != nil {
		return err
	}

	t.Set(reflect.ValueOf(v))

	return nil
}

type parser interface {
	fmt.Stringer

	parse(string, bool) (interface{}, error)
}

type valueParser struct {
	t reflect.Type
}

func (vp *valueParser) String() string { return vp.t.String() }

func (vp *valueParser) parse(v string, ptr bool) (interface{}, error) {
	rv := reflect.New(vp.t)

	if err := rv.Interface().(Value).Parse(v); err != nil {
		return nil, err
	}

	if ptr {
		return rv.Interface(), nil
	}

	return rv.Elem().Interface(), nil
}

type mapParser struct {
	t reflect.Type

	vp, kp parser

	vptr, kptr bool
}

func (mp *mapParser) String() string {
	return fmt.Sprintf("map[%s]%s", mp.kp.String(), mp.vp.String())
}

func (mp *mapParser) parse(v string, ptr bool) (interface{}, error) {
	args := strings.Split(v, ",")
	res := reflect.MakeMap(mp.t)

	for _, arg := range args {
		vs := strings.SplitN(arg, "=", 2)

		if len(vs) != 2 {
			continue
		}

		k, err := mp.kp.parse(vs[0], mp.kptr)

		if err != nil {
			return nil, err
		}

		v, err := mp.vp.parse(vs[1], mp.vptr)

		if err != nil {
			return nil, err
		}

		res.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}

	return res.Interface(), nil

}

type sliceParser struct {
	t reflect.Type

	p   parser
	ptr bool
}

func (sp *sliceParser) String() string {
	return fmt.Sprintf("[]%s", sp.p.String())
}

func (sp *sliceParser) parse(v string, ptr bool) (interface{}, error) {
	args, err := csv.NewReader(strings.NewReader(v)).Read()

	if err != nil {
		return nil, errors.Wrapf(err, "%q is not a correct slice value", v)
	}

	res := reflect.MakeSlice(sp.t, 0, len(args))

	for _, arg := range args {
		v, err := sp.p.parse(arg, sp.ptr)

		if err != nil {
			return nil, err
		}

		res = reflect.Append(res, reflect.ValueOf(v))
	}

	return res.Interface(), nil
}

type stringParser struct{}

func (stringParser) String() string { return "string" }

func (stringParser) parse(v string, ptr bool) (interface{}, error) {
	if ptr {
		x := v
		return &x, nil
	}

	return v, nil
}

type intTransformer func(int64, bool) interface{}

var intTransformers = map[reflect.Kind]intTransformer{
	reflect.Int: func(v int64, ptr bool) interface{} {
		if ptr {
			x := int(v)
			return &x
		}

		return int(v)
	},
	reflect.Int64: func(v int64, ptr bool) interface{} {
		if ptr {
			x := v
			return &x
		}

		return v
	},
}

type floatParser struct{}

func (floatParser) String() string { return "float" }

func (floatParser) parse(value string, ptr bool) (interface{}, error) {
	var v, err = strconv.ParseFloat(value, 64)

	if err != nil {
		return nil, err
	}

	if ptr {
		return &v, nil
	}

	return v, nil
}

type intParser struct {
	transformer intTransformer
}

func (*intParser) String() string { return "integer" }

func (s *intParser) parse(value string, ptr bool) (interface{}, error) {
	var v, err = strconv.ParseInt(value, 10, 0)

	if err != nil {
		return nil, err
	}

	return s.transformer(v, ptr), nil
}

type timeParser struct{}

func (timeParser) String() string { return "time" }

func (s timeParser) parse(value string, ptr bool) (interface{}, error) {
	t, err := time.Parse("2006-01-02T15:04:05", value)

	if err != nil {
		return nil, err
	}

	if ptr {
		return &t, nil
	}

	return t, nil
}

type durationParser struct{}

func (durationParser) String() string { return "duration" }

func (s durationParser) parse(value string, ptr bool) (interface{}, error) {
	d, err := time.ParseDuration(value)

	if err != nil {
		return nil, err
	}

	if ptr {
		return &d, nil
	}

	return d, nil
}
