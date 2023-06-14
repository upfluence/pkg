package ioutil

import "io"

type LimitedWriter struct {
	Writer io.Writer
	N      int

	written int
}

func (lw *LimitedWriter) Overflowed() bool {
	return lw.written >= lw.N
}

func (lw *LimitedWriter) Write(buf []byte) (int, error) {
	if lw.Overflowed() {
		return len(buf), nil
	}

	left := lw.N - lw.written
	l := len(buf)

	if left < l {
		buf = buf[:left]
	}

	n, err := lw.Writer.Write(buf)

	lw.written += n

	return l, err
}
