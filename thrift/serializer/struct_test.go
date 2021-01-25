package serializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upfluence/thrift/lib/go/thrift"
)

func TestStringSerialization(t *testing.T) {
	vs := []TStruct{&stringTStruct{"bar"}, &stringTStruct{"buz"}}

	out, err := NewDefaultTSerializer().WriteString(SliceWriter(vs))

	assert.Nil(t, err)
	assert.Equal(t, "[\"rec\",2,\"bar\",\"buz\"]", out)

	var dvs []TStruct

	err = NewDefaultTDeserializer().ReadString(
		SliceReader(func(p thrift.TProtocol) error {
			var v stringTStruct

			if err := v.Read(p); err != nil {
				return err
			}

			dvs = append(dvs, &v)
			return nil
		}),
		out,
	)

	assert.Nil(t, err)
	assert.Equal(t, vs, dvs)
}

func TestMapSerialization(t *testing.T) {
	var (
		dvs = make(map[string]TStruct)
		vs  = map[string]TStruct{
			"foo": &stringTStruct{"bar"},
			"biz": &stringTStruct{"buz"},
		}
	)

	out, err := NewDefaultTSerializer().WriteString(MapWriter(vs))
	require.NoError(t, err)

	err = NewDefaultTDeserializer().ReadString(
		MapReader(func(k string, p thrift.TProtocol) error {
			var v stringTStruct

			if err := v.Read(p); err != nil {
				return err
			}

			dvs[k] = &v
			return nil
		}),
		out,
	)

	require.NoError(t, err)
	assert.Equal(t, vs, dvs)
}
