package serializer

import (
	"github.com/upfluence/errors"

	"github.com/upfluence/thrift/lib/go/thrift"
)

var (
	ErrNotImplemented = errors.New("thrift/serializer: not implemented")
	ErrInvalidType    = errors.New("thrift/serializer: invalid type")
)

type TStruct interface {
	Write(thrift.TProtocol) error
	Read(thrift.TProtocol) error
}

type SliceWriter []TStruct

func (sw SliceWriter) Read(p thrift.TProtocol) error {
	return ErrNotImplemented
}

func (sw SliceWriter) Write(p thrift.TProtocol) error {
	if err := p.WriteListBegin(thrift.STRUCT, len(sw)); err != nil {
		return err
	}

	for _, s := range sw {
		if err := s.Write(p); err != nil {
			return err
		}
	}

	return p.WriteListEnd()
}

type SliceReader func(thrift.TProtocol) error

func (sr SliceReader) Read(p thrift.TProtocol) error {
	t, s, err := p.ReadListBegin()

	if err != nil {
		return err
	}

	if t != thrift.STRUCT {
		return ErrInvalidType
	}

	for i := 0; i < s; i++ {
		if err := sr(p); err != nil {
			return err
		}
	}

	return p.ReadListEnd()
}

func (sr SliceReader) Write(thrift.TProtocol) error {
	return ErrNotImplemented
}

type MapWriter map[string]TStruct

func (mw MapWriter) Read(p thrift.TProtocol) error {
	return ErrNotImplemented
}

func (mw MapWriter) Write(p thrift.TProtocol) error {
	if err := p.WriteMapBegin(thrift.STRING, thrift.STRUCT, len(mw)); err != nil {
		return err
	}

	for k, s := range mw {
		if err := p.WriteString(k); err != nil {
			return err
		}

		if err := s.Write(p); err != nil {
			return err
		}
	}

	return p.WriteMapEnd()
}

type MapReader func(string, thrift.TProtocol) error

func (mr MapReader) Read(p thrift.TProtocol) error {
	kt, vt, s, err := p.ReadMapBegin()

	if err != nil {
		return err
	}

	if kt != thrift.STRING || vt != thrift.STRUCT {
		return ErrInvalidType
	}

	for i := 0; i < s; i++ {
		k, err := p.ReadString()

		if err != nil {
			return err
		}

		if err := mr(k, p); err != nil {
			return err
		}
	}

	return p.ReadMapEnd()
}

func (mr MapReader) Write(thrift.TProtocol) error {
	return ErrNotImplemented
}
