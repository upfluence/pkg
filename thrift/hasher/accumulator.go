package hasher

import (
	"errors"
	"hash"
	"io"
)

type accumulator interface {
	io.Writer
	io.Closer

	beginOrdered() error
	beginUnordered() error

	endOrdered() error
	endUnordered() error
}

type fastAccumulator struct {
	io.Writer
}

func (fa *fastAccumulator) beginOrdered() error   { return nil }
func (fa *fastAccumulator) beginUnordered() error { return nil }
func (fa *fastAccumulator) endOrdered() error     { return nil }
func (fa *fastAccumulator) endUnordered() error   { return nil }
func (fa *fastAccumulator) Close() error          { return nil }

type context struct {
	hashers             []hash.Hash
	orderedCount        int
	hasUnorderedContext bool
}

type deterministicAccumulator struct {
	hasherFunc func() hash.Hash
	writer     io.Writer

	ctxs []*context
}

func newDeterministicAccumulator(w io.Writer, hasherFunc func() hash.Hash) accumulator {
	return &deterministicAccumulator{
		writer:     w,
		hasherFunc: hasherFunc,
	}
}

func (da *deterministicAccumulator) context() *context {
	if len(da.ctxs) == 0 {
		da.ctxs = append(da.ctxs, &context{})
	}

	return da.ctxs[len(da.ctxs)-1]
}

func (da *deterministicAccumulator) beginOrdered() error {
	ctx := da.context()

	if ctx.orderedCount == 0 {
		ctx.hashers = append(ctx.hashers, da.hasherFunc())
	}

	ctx.orderedCount++

	return nil
}

func (da *deterministicAccumulator) beginUnordered() error {
	da.ctxs = append(da.ctxs, &context{hasUnorderedContext: true})

	return nil
}

func (da *deterministicAccumulator) exitContext(hasher hash.Hash) error {
	da.ctxs = da.ctxs[:len(da.ctxs)-1]

	if len(da.ctxs) == 0 {
		_, err := da.writer.Write(hasher.Sum(nil))

		return err
	}

	ctx := da.ctxs[len(da.ctxs)-1]

	if ctx.orderedCount == 0 {
		ctx.hashers = append(ctx.hashers, hasher)

		return nil
	}

	_, err := ctx.hashers[len(ctx.hashers)-1].Write(hasher.Sum(nil))

	return err
}

func (da *deterministicAccumulator) endOrdered() error {
	ctx := da.context()

	if ctx.orderedCount == 0 {
		return errors.New("unexpected end of ordered context")
	}

	ctx.orderedCount--

	if ctx.orderedCount == 0 {
		if !ctx.hasUnorderedContext {
			return da.exitContext(ctx.hashers[len(ctx.hashers)-1])
		}
	}

	return nil
}

func (da *deterministicAccumulator) endUnordered() error {
	ctx := da.context()

	if ctx.orderedCount != 0 {
		return errors.New("unexpected end of unordered context")
	}

	h := da.hasherFunc()

	buf := make([]byte, h.Size())

	for _, h := range ctx.hashers {
		for i, b := range h.Sum(nil) {
			buf[i] ^= b
		}
	}

	if _, err := h.Write(buf); err != nil {
		return err
	}

	return da.exitContext(h)
}

func (da *deterministicAccumulator) Write(p []byte) (int, error) {
	if len(da.ctxs) == 0 {
		return da.writer.Write(p)
	}

	ctx := da.context()

	if ctx.orderedCount == 0 {
		ctx.hashers = append(ctx.hashers, da.hasherFunc())
	}

	return ctx.hashers[len(ctx.hashers)-1].Write(p)
}

func (da *deterministicAccumulator) Close() error {
	if len(da.ctxs) != 0 {
		return errors.New("unexpected close of accumulator with open contexts")
	}

	return nil
}
