package sql

import (
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type fakeTStruct struct {
	value int64
}

func (t fakeTStruct) Write(p thrift.TProtocol) error {
	return p.WriteI64(t.value)
}

func (t *fakeTStruct) Read(p thrift.TProtocol) error {
	val, err := p.ReadI64()

	if err != nil {
		return err
	}

	t.value = val

	return nil
}

func (t *fakeTStruct) String() string {
	return fmt.Sprint(t.value)
}

func TestNullableThrift_Scan(t *testing.T) {
	for _, tt := range []struct {
		name      string
		data      any
		wantValue *fakeTStruct
		wantErr   bool
	}{
		{
			name:      "nil value",
			data:      nil,
			wantErr:   false,
			wantValue: nil,
		},
		{
			name:      "with value",
			data:      []byte{0, 0, 0, 0, 0, 0, 0, 42},
			wantErr:   false,
			wantValue: &fakeTStruct{value: 42},
		},
		{
			name:      "invalid value",
			data:      42,
			wantErr:   true,
			wantValue: nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				s NullableThrift[fakeTStruct, *fakeTStruct]

				err = s.Scan(tt.data)
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantValue, s.Data)
		})
	}
}

func TestNullableThrift_Value(t *testing.T) {
	for _, tt := range []struct {
		name      string
		haveValue *fakeTStruct
		wantData  driver.Value
		wantErr   bool
	}{
		{
			name:      "nil value",
			haveValue: nil,
			wantData:  nil,
			wantErr:   false,
		},
		{
			name:      "with value",
			haveValue: &fakeTStruct{value: 42},
			wantData:  []byte{0, 0, 0, 0, 0, 0, 0, 42},
			wantErr:   false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				s = NullableThrift[fakeTStruct, *fakeTStruct]{
					Data: tt.haveValue,
				}

				data, err = s.Value()
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantData, data)
		})
	}
}
