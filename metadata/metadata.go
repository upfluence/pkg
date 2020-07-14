package metadata

import (
	"fmt"
	"net/textproto"
	"strings"
)

type Metadata map[string][]string

func Pairs(kvs ...string) Metadata {
	if len(kvs)%2 == 1 {
		panic(fmt.Sprintf("metadata: Pairs got the odd number of input pairs for metadata: %d", len(kvs)))
	}

	var (
		key string

		md = make(Metadata, len(kvs)/2)
	)

	for i, v := range kvs {
		if i%2 == 0 {
			key = v
			continue
		}

		md.Add(key, v)
	}

	return md
}

func (md Metadata) Set(k string, vs []string) {
	md[textproto.CanonicalMIMEHeaderKey(k)] = vs
}

func (md Metadata) Add(k, v string) { md.Set(k, []string{v}) }

func (md Metadata) Get(k string) []string {
	if len(md) == 0 {
		return nil
	}

	return md[textproto.CanonicalMIMEHeaderKey(k)]
}

func (md Metadata) Fetch(k string) string {
	return strings.Join(md.Get(k), ";")
}

func (md Metadata) Append(childMD Metadata) {
	for k, vs := range childMD {
		md.Set(k, vs)
	}
}

type Encoding interface {
	Encode(interface{}) ([]string, error)
	Decode([]string, interface{}) error
}

var DefaultEncoding Encoding = identityEncoding{}

func Encode(md Metadata, k string, v interface{}) error {
	vs, err := DefaultEncoding.Encode(v)

	if err != nil {
		return err
	}

	md.Set(k, vs)
	return nil
}

func Decode(md Metadata, k string, v interface{}) error {
	vs := md.Get(k)

	return DefaultEncoding.Decode(vs, v)
}

type identityEncoding struct{}

func (identityEncoding) Encode(v interface{}) ([]string, error) {
	var str string

	switch vv := v.(type) {
	case string:
		str = vv
	case *string:
		if vv != nil {
			str = *vv
		}
	default:
		return nil, fmt.Errorf("%T value type not handled", v)
	}

	return []string{str}, nil
}

func (identityEncoding) Decode(vs []string, v interface{}) error {
	if len(vs) != 1 {
		return fmt.Errorf("invalid number of argument")
	}

	vv, ok := v.(*string)

	if !ok {
		return fmt.Errorf("%T value type not handled", v)
	}

	*vv = vs[0]
	return nil
}
