package exporter

import "io"

type Exporter interface {
	io.Closer

	Export(<-chan bool)
}
