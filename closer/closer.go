package closer

import (
	"context"
	"io"
)

type Closer interface {
	io.Closer
}

type Shutdowner interface {
	Closer

	Shutdown(context.Context) error
}
