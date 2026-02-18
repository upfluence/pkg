package hasher

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"io"
	"sync"

	"github.com/upfluence/thrift/lib/go/thrift"
)

type Hasher64 struct {
	hasherPool sync.Pool

	accumulatorFunc func(io.Writer, func() hash.Hash) accumulator
}

func NewFNVHasher64() *Hasher64 { return NewHasher64(fnv.New64) }

func NewHasher64(fn func() hash.Hash64) *Hasher64 {
	return &Hasher64{
		hasherPool:      sync.Pool{New: func() any { return fn() }},
		accumulatorFunc: newDeterministicAccumulator,
	}
}

func newHasher64(hfn func() hash.Hash64, afn func(io.Writer, func() hash.Hash) accumulator) *Hasher64 {
	return &Hasher64{
		hasherPool:      sync.Pool{New: func() any { return hfn() }},
		accumulatorFunc: afn,
	}
}

type protocol struct {
	thrift.TProtocol

	acc accumulator
}

func (p *protocol) WriteStructBegin(name string) error {
	return p.acc.beginUnordered()
}

func (p *protocol) WriteStructEnd() error {
	return p.acc.endUnordered()
}

func (p *protocol) WriteFieldBegin(name string, tid thrift.TType, id int16) error {
	if err := p.acc.beginOrdered(); err != nil {
		return err
	}

	if err := p.WriteByte(byte(tid)); err != nil {
		return err
	}

	return p.WriteI16(id)
}

func (p *protocol) WriteFieldEnd() error {
	return p.acc.endOrdered()
}

func (p *protocol) WriteFieldStop() error {
	return nil
}

func (p *protocol) beginContainer(size int, ordered bool, ids ...thrift.TType) error {
	if err := p.acc.beginOrdered(); err != nil {
		return err
	}

	for _, id := range ids {
		if err := p.WriteByte(byte(id)); err != nil {
			return err
		}
	}

	if err := p.WriteI64(int64(size)); err != nil {
		return err
	}

	if ordered {
		return p.acc.beginOrdered()
	} else {
		return p.acc.beginUnordered()
	}
}

func (p *protocol) endContainer(ordered bool) error {
	var err error

	if ordered {
		err = p.acc.endOrdered()
	} else {
		err = p.acc.endUnordered()
	}

	if err != nil {
		return err
	}

	return p.acc.endOrdered()
}

func (p *protocol) WriteMapBegin(kid, vid thrift.TType, size int) error {
	return p.beginContainer(size, false, kid, vid)
}

func (p *protocol) WriteMapEnd() error {
	return p.endContainer(false)
}
func (p *protocol) WriteListBegin(id thrift.TType, size int) error {
	return p.beginContainer(size, true, id)
}

func (p *protocol) WriteListEnd() error {
	return p.endContainer(true)
}

func (p *protocol) WriteSetBegin(id thrift.TType, size int) error {
	return p.beginContainer(size, false, id)
}

func (p *protocol) WriteSetEnd() error {
	return p.endContainer(false)
}

func (p *protocol) WriteBool(value bool) error {
	var b int16

	if value {
		b = 1
	}

	return p.WriteI16(b)
}

func (p *protocol) WriteByte(value byte) error {
	return p.WriteI16(int16(value))
}

func (p *protocol) WriteI16(value int16) error {
	return binary.Write(p.acc, binary.BigEndian, value)
}

func (p *protocol) WriteI32(value int32) error {
	return binary.Write(p.acc, binary.BigEndian, value)
}

func (p *protocol) WriteI64(value int64) error {
	return binary.Write(p.acc, binary.BigEndian, value)
}

func (p *protocol) WriteDouble(value float64) error {
	return binary.Write(p.acc, binary.BigEndian, value)
}
func (p *protocol) WriteString(value string) error {
	_, err := io.WriteString(p.acc, value)

	return err
}

func (p *protocol) WriteBinary(value []byte) error {
	_, err := p.acc.Write(value)

	return err
}

func (h *Hasher64) Hash(msg thrift.TStruct) (uint64, error) {
	hh := h.hasherPool.Get().(hash.Hash64)

	defer func() {
		hh.Reset()
		h.hasherPool.Put(hh)
	}()

	acc := h.accumulatorFunc(hh, func() hash.Hash { return h.hasherPool.Get().(hash.Hash) })

	if err := msg.Write(&protocol{acc: acc}); err != nil {
		return 0, err
	}

	if err := acc.Close(); err != nil {
		return 0, err
	}

	return hh.Sum64(), nil
}
