package walker

import (
	"errors"
	"reflect"
	"unicode"
)

var (
	SkipStruct            = errors.New("skip struct")
	ErrShouldBeAStructPtr = errors.New("cfg: input should be a pointer")
)

type Field struct {
	Field reflect.StructField
	Value reflect.Value

	Ancestor *Field
}

type WalkFunc func(*Field) error

func Walk(in interface{}, fn WalkFunc) error {
	if in == nil {
		return ErrShouldBeAStructPtr
	}

	inv := reflect.ValueOf(in)

	if inv.Type().Kind() != reflect.Ptr {
		return ErrShouldBeAStructPtr
	}

	if inv.Type().Elem().Kind() != reflect.Struct {
		return ErrShouldBeAStructPtr
	}

	if inv.IsNil() {
		return ErrShouldBeAStructPtr
	}

	return walk(inv, fn, nil)
}

func indirectedType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}

	return t
}

func indirectedValue(v reflect.Value) reflect.Value {
	if v.Type().Kind() == reflect.Ptr {
		return v.Elem()
	}

	return v
}

func addressValue(v reflect.Value) reflect.Value {
	if v.Type().Kind() == reflect.Ptr {
		return v
	}

	return v.Addr()
}

func walk(v reflect.Value, fn WalkFunc, a *Field) error {
	vit := indirectedType(v.Type())

	for i := 0; i < vit.NumField(); i++ {
		sf := vit.Field(i)
		f := Field{
			Field:    sf,
			Value:    addressValue(v),
			Ancestor: a,
		}

		nv := indirectedValue(v).FieldByName(sf.Name)

		if sf.Type.Kind() != reflect.Ptr {
			nv = nv.Addr()
		} else if !nv.CanSet() {
			continue
		}

		if unicode.IsUpper(rune(sf.Name[0])) {
			switch err := fn(&f); err {
			case SkipStruct:
				continue
			case nil:
			default:
				return err
			}
		}

		if indirectedType(sf.Type).Kind() != reflect.Struct {
			continue
		}

		if sf.Type.Kind() == reflect.Ptr && nv.IsNil() {
			nv.Set(reflect.New(sf.Type.Elem()))
		}

		if err := walk(nv, fn, &f); err != nil {
			return err
		}
	}

	return nil
}
