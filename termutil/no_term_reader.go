package termutil

import (
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

type NoTermReader struct {
	Reader io.Reader

	once   sync.Once
	isTerm bool
}

func (ntr *NoTermReader) Read(p []byte) (int, error) {
	ntr.once.Do(func() { ntr.isTerm = isTerm(ntr.Reader) })

	if ntr.isTerm {
		return 0, io.EOF
	}

	return ntr.Reader.Read(p)
}

func isTerm(r io.Reader) bool {
	f, ok := r.(*os.File)

	return ok && term.IsTerminal(int(f.Fd()))
}
