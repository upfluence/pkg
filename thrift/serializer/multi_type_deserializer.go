package serializer

import (
	"errors"
	"io"
	"strings"

	"github.com/upfluence/pkg/encoding"
	"github.com/upfluence/pkg/thrift/thriftutil"
)

var (
	ErrProtocolNotProvided = errors.New("thrift/serializer: Protocol not provided")
	ErrEncodingNotProvided = errors.New("thrift/serializer: Encoding not provided")
)

type TMultiTypeDeserializer struct {
	*TDeserializer

	pfs map[string]thriftutil.TTypedProtocolFactory
	es  map[string]encoding.Encoding
}

func NewDefaultTMultiTypeDeserializer() *TMultiTypeDeserializer {
	return NewTMultiTypeDeserializer(
		thriftutil.JSONProtocolFactory,
		[]thriftutil.TTypedProtocolFactory{thriftutil.BinaryProtocolFactory},
		[]encoding.Encoding{
			encoding.SnappyEncoding,
			encoding.GZipEncoding,
			encoding.Base64Encoding,
		},
	)
}

func NewTMultiTypeDeserializer(dpf thriftutil.TTypedProtocolFactory, pfs []thriftutil.TTypedProtocolFactory, es []encoding.Encoding) *TMultiTypeDeserializer {
	pfactories := make(map[string]thriftutil.TTypedProtocolFactory)
	encs := make(map[string]encoding.Encoding)

	for _, pf := range append(pfs, dpf) {
		pfactories[pf.ContentType()] = pf
	}

	for _, e := range es {
		encs[e.ContentType()] = e
	}

	return &TMultiTypeDeserializer{
		TDeserializer: NewTDeserializer(dpf),
		pfs:           pfactories,
		es:            encs,
	}
}

func (mtd *TMultiTypeDeserializer) deserializer(ct string) (*TDeserializer, error) {
	if ct == "" {
		return mtd.TDeserializer, nil
	}

	var (
		fs = strings.Split(ct, "+")

		es []encoding.Encoding
	)

	pf, ok := mtd.pfs[fs[0]]

	if !ok {
		return nil, ErrProtocolNotProvided
	}

	if len(fs) > 1 {
		for _, f := range fs[1:] {
			e, ok := mtd.es[f]

			if !ok {
				return nil, ErrEncodingNotProvided
			}

			es = append(es, e)
		}
	}

	return NewTDeserializer(pf, es...), nil
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
