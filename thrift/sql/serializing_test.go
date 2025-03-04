package sql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upfluence/errors"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/thrift/serializer"
	"github.com/upfluence/pkg/thrift/thriftutil"
)

type fakeTStruct struct {
	value int64
}

func (t fakeTStruct) Write(p thrift.TProtocol) error {
	return errors.WithStack(p.WriteI64(t.value))
}

func (t *fakeTStruct) Read(p thrift.TProtocol) error {
	val, err := p.ReadI64()

	if err != nil {
		return errors.WithStack(err)
	}

	t.value = val

	return nil
}

func (t *fakeTStruct) String() string {
	return fmt.Sprint(t.value)
}

func TestNullableThrift_Scan(t *testing.T) {
	for _, tt := range []struct {
		name              string
		fakeValue         *fakeTStruct
		serializerFactory *serializer.TSerializerFactory
	}{
		{
			name:      "nil value",
			fakeValue: nil,
		},
		{
			name: "with value",
			fakeValue: &fakeTStruct{
				value: 42,
			},
		},
		{
			name: "with custom serializer factory",
			fakeValue: &fakeTStruct{
				value: 42,
			},
			serializerFactory: serializer.NewTSerializerFactory(
				thriftutil.JSONProtocolFactory,
			),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				s = NullThrift[
					fakeTStruct,
					*fakeTStruct,
				]{
					Data:              tt.fakeValue,
					SerializerFactory: tt.serializerFactory,
				}

				data, err = s.Value()
			)

			require.NoError(t, err)

			s.Data = nil

			require.NoError(t, s.Scan(data))
			assert.Equal(t, tt.fakeValue, s.Data)
		})
	}
}
