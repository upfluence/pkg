package hasher

import (
	"hash"
	"hash/fnv"
	"sync"

	"github.com/upfluence/pkg/v2/thrift/serializer"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type Hasher64 struct {
	hasherPool sync.Pool
	serializer serializer.Serializer
}

func NewFNVHasher64() *Hasher64 { return NewHasher64(fnv.New64) }

func NewHasher64(fn func() hash.Hash64) *Hasher64 {
	return &Hasher64{
		hasherPool: sync.Pool{New: func() any { return fn() }},
		serializer: serializer.NewTSerializer(thrift.NewTBinaryProtocolFactoryDefault()),
	}
}

func (h *Hasher64) Hash(msg serializer.TStruct) (uint64, error) {
	hh := h.hasherPool.Get().(hash.Hash64)

	defer func() {
		hh.Reset()
		h.hasherPool.Put(hh)
	}()

	if err := h.serializer.WriteTo(msg, hh); err != nil {
		return 0, err
	}

	return hh.Sum64(), nil
}
