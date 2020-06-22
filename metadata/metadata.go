package metadata

import (
	"fmt"
	"net/textproto"
)

type Metadata map[string][]string

func (md Metadata) Set(k string, vs []string) {
	md[textproto.CanonicalMIMEHeaderKey(k)] = vs
}

func (md Metadata) Get(k string) []string {
	if len(md) == 0 {
		return nil
	}

	return md[textproto.CanonicalMIMEHeaderKey(k)]
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
