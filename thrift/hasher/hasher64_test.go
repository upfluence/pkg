package hasher

import (
	"hash"
	"hash/fnv"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upfluence/thrift/lib/go/thrift"
)

type notReadable struct{}

func (notReadable) Read(thrift.TProtocol) error { panic("unexpected") }
func (notReadable) String() string              { return "test" }

// metadata mimics core.Metadata: a struct holding a map<string, list<string>>.
type metadata map[string][]string

func (metadata) Read(thrift.TProtocol) error { panic("unexpected") }
func (metadata) String() string              { return "metadata" }

func (m metadata) Write(p thrift.TProtocol) error {
	if err := p.WriteStructBegin("Metadata"); err != nil {
		return err
	}

	if err := p.WriteFieldBegin("entries", thrift.MAP, 1); err != nil {
		return err
	}

	if err := p.WriteMapBegin(thrift.STRING, thrift.LIST, len(m)); err != nil {
		return err
	}

	for k, vs := range m {
		if err := p.WriteString(k); err != nil {
			return err
		}

		if err := p.WriteListBegin(thrift.STRING, len(vs)); err != nil {
			return err
		}

		for _, v := range vs {
			if err := p.WriteString(v); err != nil {
				return err
			}
		}

		if err := p.WriteListEnd(); err != nil {
			return err
		}
	}

	if err := p.WriteMapEnd(); err != nil {
		return err
	}

	if err := p.WriteFieldEnd(); err != nil {
		return err
	}

	if err := p.WriteFieldStop(); err != nil {
		return err
	}

	return p.WriteStructEnd()
}

// pair is a struct with two string fields.
type pair [2]string

func (pair) Read(thrift.TProtocol) error { panic("unexpected") }
func (pair) String() string              { return "pair" }

func (pa pair) Write(p thrift.TProtocol) error {
	if err := p.WriteStructBegin("Pair"); err != nil {
		return err
	}

	for i, s := range pa {
		if err := p.WriteFieldBegin("f", thrift.STRING, int16(i+1)); err != nil {
			return err
		}

		if err := p.WriteString(s); err != nil {
			return err
		}

		if err := p.WriteFieldEnd(); err != nil {
			return err
		}
	}

	if err := p.WriteFieldStop(); err != nil {
		return err
	}

	return p.WriteStructEnd()
}

type stringSet []string

func (stringSet) Read(thrift.TProtocol) error { panic("unexpected") }
func (stringSet) String() string              { return "set" }

func (ss stringSet) Write(p thrift.TProtocol) error {
	if err := p.WriteSetBegin(thrift.STRING, len(ss)); err != nil {
		return err
	}

	for _, s := range ss {
		if err := p.WriteString(s); err != nil {
			return err
		}
	}

	return p.WriteSetEnd()
}

type unbalanced struct{ notReadable }

func (unbalanced) Write(p thrift.TProtocol) error { return p.WriteListEnd() }

type unclosed struct{ notReadable }

func (unclosed) Write(p thrift.TProtocol) error {
	return p.WriteListBegin(thrift.STRING, 0)
}

func TestHasher64Distinguishes(t *testing.T) {
	for _, tt := range []struct {
		name string
		a, b thrift.TStruct
	}{
		{
			name: "values swapped between map keys",
			a:    metadata{"alpha": {"one-x"}, "beta": {"two-y"}},
			b:    metadata{"alpha": {"two-y"}, "beta": {"one-x"}},
		},
		{
			name: "key and value swapped",
			a:    metadata{"alpha": {"one-x"}},
			b:    metadata{"one-x": {"alpha"}},
		},
		{
			name: "element boundary moved",
			a:    metadata{"alpha": {"ab", "c"}},
			b:    metadata{"alpha": {"a", "bc"}},
		},
		{
			name: "element order",
			a:    metadata{"alpha": {"one", "two"}},
			b:    metadata{"alpha": {"two", "one"}},
		},
		{
			name: "last bytes swapped between struct fields",
			a:    pair{"api-1", "10.0.0.2"},
			b:    pair{"api-2", "10.0.0.1"},
		},
		{
			name: "single characters swapped between struct fields",
			a:    pair{"a", "1"},
			b:    pair{"1", "a"},
		},
		{
			name: "repeated set elements do not cancel out",
			a:    stringSet{"x", "x", "y", "y"},
			b:    stringSet{"z", "z", "w", "w"},
		},
		{
			name: "set element removed",
			a:    stringSet{"x", "y", "z"},
			b:    stringSet{"x", "y"},
		},
	} {
		for hname, fn := range map[string]func() hash.Hash64{
			"fnv1a": fnv.New64a,
			"fnv1":  fnv.New64,
		} {
			t.Run(tt.name+"/"+hname, func(t *testing.T) {
				h := NewHasher64(fn)

				a, err := h.Hash(tt.a)

				require.NoError(t, err)

				b, err := h.Hash(tt.b)

				require.NoError(t, err)

				assert.NotEqual(t, a, b)
			})
		}
	}
}

// entry is a map entry of orderedMetadata.
type entry struct {
	key    string
	values []string
}

// orderedMetadata is a metadata whose entries are written in the given order,
// unlike a go map whose iteration order is random.
type orderedMetadata []entry

func (orderedMetadata) Read(thrift.TProtocol) error { panic("unexpected") }
func (orderedMetadata) String() string              { return "ordered_metadata" }

func (om orderedMetadata) Write(p thrift.TProtocol) error {
	if err := p.WriteStructBegin("Metadata"); err != nil {
		return err
	}

	if err := p.WriteFieldBegin("entries", thrift.MAP, 1); err != nil {
		return err
	}

	if err := p.WriteMapBegin(thrift.STRING, thrift.LIST, len(om)); err != nil {
		return err
	}

	for _, e := range om {
		if err := p.WriteString(e.key); err != nil {
			return err
		}

		if err := p.WriteListBegin(thrift.STRING, len(e.values)); err != nil {
			return err
		}

		for _, v := range e.values {
			if err := p.WriteString(v); err != nil {
				return err
			}
		}

		if err := p.WriteListEnd(); err != nil {
			return err
		}
	}

	if err := p.WriteMapEnd(); err != nil {
		return err
	}

	if err := p.WriteFieldEnd(); err != nil {
		return err
	}

	if err := p.WriteFieldStop(); err != nil {
		return err
	}

	return p.WriteStructEnd()
}

func permutations[T any](vs []T) [][]T {
	if len(vs) <= 1 {
		return [][]T{slices.Clone(vs)}
	}

	var res [][]T

	for i := range vs {
		rest := slices.Concat(vs[:i:i], vs[i+1:])

		for _, p := range permutations(rest) {
			res = append(res, append([]T{vs[i]}, p...))
		}
	}

	return res
}

func TestHasher64IsOrderIndependent(t *testing.T) {
	h := NewFNVHasher64()

	for _, tt := range []struct {
		name string
		ins  []thrift.TStruct
	}{
		{
			name: "set",
			ins: func() []thrift.TStruct {
				var res []thrift.TStruct

				for _, p := range permutations([]string{"a", "b", "c", "d", "e"}) {
					res = append(res, stringSet(p))
				}

				return res
			}(),
		},
		{
			name: "map entries",
			ins: func() []thrift.TStruct {
				var res []thrift.TStruct

				for _, p := range permutations([]entry{
					{key: "a", values: []string{"1", "2"}},
					{key: "b", values: []string{"3"}},
					{key: "c"},
					{key: "d", values: []string{"4", "5", "6"}},
				}) {
					res = append(res, orderedMetadata(p))
				}

				return res
			}(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Greater(t, len(tt.ins), 1)

			want, err := h.Hash(tt.ins[0])

			require.NoError(t, err)

			for _, in := range tt.ins[1:] {
				got, err := h.Hash(in)

				require.NoError(t, err)
				assert.Equal(t, want, got, "%v", in)
			}
		})
	}
}

func TestHasher64ListOrderMatters(t *testing.T) {
	h := NewFNVHasher64()

	hashes := make(map[uint64]struct{})

	for _, p := range permutations([]string{"a", "b", "c"}) {
		res, err := h.Hash(orderedMetadata{{key: "k", values: p}})

		require.NoError(t, err)

		hashes[res] = struct{}{}
	}

	assert.Len(t, hashes, 6)
}

func TestHasher64Unbalanced(t *testing.T) {
	h := NewFNVHasher64()

	for name, in := range map[string]thrift.TStruct{
		"end without begin": unbalanced{},
		"begin without end": unclosed{},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := h.Hash(in)

			assert.ErrorIs(t, err, errUnbalanced)
		})
	}
}

func TestHasher64IsStable(t *testing.T) {
	// Pinned so that a change of the scheme, which would change every hash
	// derived from it, is a conscious decision.
	got, err := NewFNVHasher64().Hash(
		metadata{"alpha": {"one", "two"}, "beta": {"three"}},
	)

	require.NoError(t, err)
	assert.Equal(t, uint64(0x15979cf50e9daa43), got)
}
