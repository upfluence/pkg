package thriftutil

import (
	"github.com/upfluence/pkg/bytesutil"
	"github.com/upfluence/thrift/lib/go/thrift"
)

var (
	transportFactory = BinaryProtocolFactory
	bufferPool       = bytesutil.NewBufferPool()
)

func Clone[T any, PT interface {
	thrift.TStruct
	*T
}](in PT) (PT, error) {
	var (
		cpy PT = new(T)
		buf    = bufferPool.Get()
	)

	defer bufferPool.Put(buf)

	if err := in.Write(transportFactory.GetProtocol(WrapWriter(buf))); err != nil {
		return nil, err
	}

	if err := cpy.Read(transportFactory.GetProtocol(WrapReader(buf))); err != nil {
		return nil, err
	}

	return cpy, nil
}

func MustClone[T any, PT interface {
	thrift.TStruct
	*T
}](in PT) PT {
	res, err := Clone(in)

	if err != nil {
		panic(err)
	}

	return res
}
