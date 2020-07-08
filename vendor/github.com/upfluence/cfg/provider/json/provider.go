package json

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/upfluence/cfg/provider"
)

var ErrJSONMalformated = errors.New("cfg/provider/json: Payload not formatted correctly")

type Provider struct {
	store map[string]interface{}
}

func NewProviderFromReader(r io.Reader) provider.Provider {
	var v = make(map[string]interface{})

	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return provider.ProvideError("json", err)
	}

	return &Provider{store: v}
}

func (*Provider) StructTag() string { return "json" }

func (p *Provider) Provide(_ context.Context, v string) (string, bool, error) {
	var (
		cur         = p.store
		splittedKey = strings.Split(v, ".")

		res interface{}
	)

	for i, k := range splittedKey {
		t := cur[k]

		if t == nil {
			return "", false, nil
		}

		if i == len(splittedKey)-1 {
			res = t
			continue
		}

		res, ok := t.(map[string]interface{})

		if !ok {
			return "", false, ErrJSONMalformated
		}

		cur = res
	}

	return stringifyValue(res), true, nil
}

func stringifyValue(v interface{}) string {
	vv := reflect.ValueOf(v)

	switch vv.Kind() {
	case reflect.Slice:
		var vs []string

		for i := 0; i < vv.Len(); i++ {
			vs = append(vs, stringifyValue(vv.Index(i).Interface()))
		}

		var b strings.Builder

		w := csv.NewWriter(&b)

		if err := w.Write(vs); err != nil {
			return strings.Join(vs, ",")
		}

		w.Flush()

		if res := b.String(); len(res) > 0 {
			return res[:len(res)-1]
		}

		return strings.Join(vs, ",")
	case reflect.Map:
		var vs []string

		for _, mkv := range vv.MapKeys() {
			mvv := vv.MapIndex(mkv)

			vs = append(
				vs,
				fmt.Sprintf(
					"%s=%s",
					stringifyValue(mkv.Interface()),
					stringifyValue(mvv.Interface()),
				),
			)
		}

		return strings.Join(vs, ",")
	}

	return fmt.Sprintf("%v", v)
}
