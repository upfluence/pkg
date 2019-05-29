package serializer

import (
	"errors"
	"io"

	"github.com/upfluence/pkg/compress"
	"github.com/upfluence/pkg/thrift/thriftutil"
)

var ErrContentTypeNotProvided = errors.New("thrift/serializer: Content type not provided")

type TMultiTypeDeserializer struct {
	*TDeserializer

	ds map[string]*TDeserializer
}

func NewDefaultTMultiTypeDeserializer() *TMultiTypeDeserializer {
	return NewTMultiTypeDeserializer(
		thriftutil.JSONProtocolFactory,
		[]thriftutil.TTypedProtocolFactory{thriftutil.BinaryProtocolFactory},
		[]compress.Compressor{compress.SnappyCompressor, compress.GZipCompressor},
	)
}

func NewTMultiTypeDeserializer(dpf thriftutil.TTypedProtocolFactory, pfs []thriftutil.TTypedProtocolFactory, cs []compress.Compressor) *TMultiTypeDeserializer {
	ds := make(map[string]*TDeserializer)

	for _, pf := range append(pfs, dpf) {
		d := NewTDeserializer(pf)
		ds[d.ContentType()] = d

		for _, c := range cs {
			d := NewTDeserializer(pf, c)
			ds[d.ContentType()] = d
		}
	}

	return &TMultiTypeDeserializer{TDeserializer: NewTDeserializer(dpf), ds: ds}
}

func (mtd *TMultiTypeDeserializer) deserializer(ct string) (*TDeserializer, error) {
	if ct == "" {
		return mtd.TDeserializer, nil
	}

	d, ok := mtd.ds[ct]

	if !ok {
		return nil, ErrContentTypeNotProvided
	}

	return d, nil
}

func (mtd *TMultiTypeDeserializer) ReadFrom(msg TStruct, r io.Reader, ct string) error {
	var d, err = mtd.deserializer(ct)

	if err != nil {
		return err
	}

	return d.ReadFrom(msg, r)
}

func (mtd *TMultiTypeDeserializer) Read(msg TStruct, p []byte, ct string) error {
	var d, err = mtd.deserializer(ct)

	if err != nil {
		return err
	}

	return d.Read(msg, p)
}

func (mtd *TMultiTypeDeserializer) ReadString(msg TStruct, p, ct string) error {
	var d, err = mtd.deserializer(ct)

	if err != nil {
		return err
	}

	return d.ReadString(msg, p)
}
