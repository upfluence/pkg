package hasher

import (
	"bytes"
	"encoding/binary"
	"hash"
	"hash/fnv"
	"math"
	"slices"
	"sync"

	"github.com/upfluence/errors"
	"github.com/upfluence/thrift/lib/go/thrift"
)

var errUnbalanced = errors.New("thrift/hasher: unbalanced thrift write calls")

// Hasher64 hashes thrift structs into a 64 bit value that only depends
// on their content: the same value is produced across processes, replicas and
// map/set iteration orders.
//
// The hash has the following properties:
//
//   - map keys and values are hashed together as one entry, so moving a value
//     from a key to another changes the hash;
//   - strings and binaries are length prefixed, so moving the boundary between
//     two consecutive elements changes the hash;
//   - the digests of unordered elements (set elements, map entries) are sorted
//     rather than xored, so identical elements do not cancel out;
//   - struct fields are hashed in the order they are written, with their ids.
//
// The output of versions before this scheme (xor of per-element hashes) is not
// compatible: hashes persisted with it must be recomputed.
//
// Use a hash with a good avalanche, such as FNV-1a. FNV-1 mixes the last byte
// of its input poorly.
type Hasher64 struct {
	hasherPool sync.Pool
}

// NewFNVHasher64 returns a Hasher64 based on FNV-1a.
func NewFNVHasher64() *Hasher64 {
	return NewHasher64(fnv.New64a)
}

func NewHasher64(fn func() hash.Hash64) *Hasher64 {
	return &Hasher64{
		hasherPool: sync.Pool{New: func() any { return fn() }},
	}
}

func (h *Hasher64) sum(b []byte) uint64 {
	hh := h.hasherPool.Get().(hash.Hash64)

	defer func() {
		hh.Reset()
		h.hasherPool.Put(hh)
	}()

	// hash.Hash writes never return an error
	_, _ = hh.Write(b)

	return hh.Sum64()
}

func (h *Hasher64) Hash(msg thrift.TStruct) (uint64, error) {
	p := &stableProtocol{h: h, stack: []*stableFrame{{kind: frameOrdered}}}

	if err := msg.Write(p); err != nil {
		return 0, errors.Wrap(err, "can not write the struct")
	}

	if len(p.stack) != 1 {
		return 0, errUnbalanced
	}

	return h.sum(p.stack[0].buf.Bytes()), nil
}

type frameKind int

const (
	// frameOrdered covers structs, lists and the root: the bytes written are
	// kept as is.
	frameOrdered frameKind = iota
	// frameSet and frameMap hash every element (or key/value entry)
	// separately, then sort the digests.
	frameSet
	frameMap
)

type stableFrame struct {
	kind frameKind

	buf bytes.Buffer

	// cur is the element being written, only used by set and map frames
	cur     bytes.Buffer
	pending int
	digests []uint64

	// header is written before the content when the frame ends
	header []byte
}

type stableProtocol struct {
	thrift.TProtocol

	h     *Hasher64
	stack []*stableFrame
}

func (p *stableProtocol) top() *stableFrame { return p.stack[len(p.stack)-1] }

func (p *stableProtocol) write(b []byte) {
	f := p.top()

	switch f.kind {
	case frameOrdered:
		f.buf.Write(b)
	default:
		f.cur.Write(b)
	}
}

// elementDone is called once a primitive was written or a container ended in
// the current frame.
func (p *stableProtocol) elementDone() {
	f := p.top()

	switch f.kind {
	case frameSet:
		f.digests = append(f.digests, p.h.sum(f.cur.Bytes()))
		f.cur.Reset()
	case frameMap:
		f.pending++

		if f.pending == 2 {
			f.digests = append(f.digests, p.h.sum(f.cur.Bytes()))
			f.cur.Reset()
			f.pending = 0
		}
	}
}

func (p *stableProtocol) primitive(b []byte) error {
	p.write(b)
	p.elementDone()

	return nil
}

func (p *stableProtocol) begin(kind frameKind, header []byte) error {
	p.stack = append(p.stack, &stableFrame{kind: kind, header: header})

	return nil
}

func (p *stableProtocol) end(kind frameKind) error {
	if len(p.stack) < 2 || p.top().kind != kind {
		return errUnbalanced
	}

	f := p.top()
	p.stack = p.stack[:len(p.stack)-1]

	var out bytes.Buffer

	out.Write(f.header)

	switch kind {
	case frameOrdered:
		out.Write(f.buf.Bytes())
	default:
		slices.Sort(f.digests)

		for _, d := range f.digests {
			out.Write(binary.BigEndian.AppendUint64(nil, d))
		}
	}

	return p.primitive(out.Bytes())
}

func (p *stableProtocol) WriteStructBegin(string) error {
	return p.begin(frameOrdered, nil)
}

func (p *stableProtocol) WriteStructEnd() error { return p.end(frameOrdered) }

func (p *stableProtocol) WriteFieldBegin(_ string, tid thrift.TType, id int16) error {
	p.write(binary.BigEndian.AppendUint16([]byte{byte(tid)}, uint16(id))) //nolint:gosec // bit pattern

	return nil
}

func (p *stableProtocol) WriteFieldEnd() error { return nil }

func (p *stableProtocol) WriteFieldStop() error {
	p.write([]byte{byte(thrift.STOP)})

	return nil
}

func containerHeader(size int, ids ...thrift.TType) []byte {
	b := make([]byte, 0, len(ids)+8)

	for _, id := range ids {
		b = append(b, byte(id))
	}

	return binary.BigEndian.AppendUint64(b, uint64(size)) //nolint:gosec // never negative
}

func (p *stableProtocol) WriteListBegin(id thrift.TType, size int) error {
	return p.begin(frameOrdered, containerHeader(size, id))
}

func (p *stableProtocol) WriteListEnd() error { return p.end(frameOrdered) }

func (p *stableProtocol) WriteSetBegin(id thrift.TType, size int) error {
	return p.begin(frameSet, containerHeader(size, id))
}

func (p *stableProtocol) WriteSetEnd() error { return p.end(frameSet) }

func (p *stableProtocol) WriteMapBegin(kid, vid thrift.TType, size int) error {
	return p.begin(frameMap, containerHeader(size, kid, vid))
}

func (p *stableProtocol) WriteMapEnd() error { return p.end(frameMap) }

func (p *stableProtocol) WriteBool(v bool) error {
	if v {
		return p.primitive([]byte{1})
	}

	return p.primitive([]byte{0})
}

func (p *stableProtocol) WriteByte(v byte) error { return p.primitive([]byte{v}) }

func (p *stableProtocol) WriteI16(v int16) error {
	return p.primitive(binary.BigEndian.AppendUint16(nil, uint16(v))) //nolint:gosec // bit pattern
}

func (p *stableProtocol) WriteI32(v int32) error {
	return p.primitive(binary.BigEndian.AppendUint32(nil, uint32(v))) //nolint:gosec // bit pattern
}

func (p *stableProtocol) WriteI64(v int64) error {
	return p.primitive(binary.BigEndian.AppendUint64(nil, uint64(v))) //nolint:gosec // bit pattern
}

func (p *stableProtocol) WriteDouble(v float64) error {
	return p.primitive(binary.BigEndian.AppendUint64(nil, math.Float64bits(v)))
}

func (p *stableProtocol) WriteString(v string) error {
	return p.primitive(append(binary.AppendUvarint(nil, uint64(len(v))), v...))
}

func (p *stableProtocol) WriteBinary(v []byte) error {
	return p.primitive(append(binary.AppendUvarint(nil, uint64(len(v))), v...))
}
