package sql

import (
	"database/sql/driver"

	"github.com/upfluence/errors"
	"github.com/upfluence/thrift/lib/go/thrift"

	"github.com/upfluence/pkg/thrift/serializer"
	"github.com/upfluence/pkg/thrift/thriftutil"
)

var defaultSerializerFactory = serializer.NewTSerializerFactory(
	thriftutil.BinaryProtocolFactory,
)

type TStructPtr[T any] interface {
	thrift.TStruct
	*T
}

type NullThrift[T any, PT TStructPtr[T]] struct {
	Data              PT
	SerializerFactory *serializer.TSerializerFactory
}

func (t NullThrift[T, PT]) serializerFactory() *serializer.TSerializerFactory {
	if t.SerializerFactory == nil {
		return defaultSerializerFactory
	}

	return t.SerializerFactory
}

func (t *NullThrift[T, PT]) Scan(value any) error {
	if value == nil {
		t.Data = nil

		return nil
	}

	data, ok := value.([]byte)

	if !ok {
		return errors.New("invalid type: expected []byte")
	}

	if t.Data == nil {
		t.Data = new(T)
	}

	return errors.WithStack(
		t.serializerFactory().GetDeserializer().Read(t.Data, data),
	)
}

func (t NullThrift[T, PT]) Value() (driver.Value, error) {
	if t.Data == nil {
		return nil, nil // nolint:nilnil
	}

	data, err := t.serializerFactory().GetSerializer().Write(t.Data)

	return data, errors.WithStack(err)
}
