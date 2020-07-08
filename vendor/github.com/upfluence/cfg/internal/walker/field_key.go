package walker

import (
	"reflect"
	"strings"
)

func walkFields(f *Field, fn func(reflect.StructField)) {
	var (
		fs = []reflect.StructField{f.Field}
		a  = f.Ancestor
	)

	for a != nil {
		fs = append(fs, a.Field)
		a = a.Ancestor
	}

	for i := len(fs); i > 0; i-- {
		fn(fs[i-1])
	}
}

func buildStructFieldKey(t string, sf reflect.StructField) []string {
	if t != "" {
		if v, ok := sf.Tag.Lookup(t); ok {
			return strings.Split(v, ",")
		}
	}

	if sf.Anonymous {
		return nil
	}

	return []string{sf.Name}
}

func BuildFieldKeys(t string, f *Field) []string {
	var fss [][]string

	walkFields(f, func(sf reflect.StructField) {
		if fs := buildStructFieldKey(t, sf); len(fs) > 0 {
			fss = append(fss, fs)
		}
	})

	if len(fss) == 0 {
		return []string{"config"}
	}

	return joinPermutation(fss, ".")
}

func joinPermutation(fss [][]string, delim string) []string {
	switch len(fss) {
	case 0:
		return nil
	case 1:
		return fss[0]
	}

	left := fss[0]
	right := joinPermutation(fss[1:], delim)

	var res []string

	for _, l := range left {
		for _, r := range right {
			res = append(res, strings.Join([]string{l, r}, delim))
		}
	}

	return res
}
