package compress

import (
	"compress/gzip"
	"io"
	"strings"

	"github.com/golang/snappy"
)

var (
	GZipCompressor = &StaticCompressor{
		Type:       "gzip",
		WriterFunc: func(w io.Writer) (io.Writer, error) { return gzip.NewWriter(w), nil },
		ReaderFunc: func(r io.Reader) (io.Reader, error) { return gzip.NewReader(r) },
	}

	SnappyCompressor = &StaticCompressor{
		Type:       "snappy",
		WriterFunc: func(w io.Writer) (io.Writer, error) { return snappy.NewWriter(w), nil },
		ReaderFunc: func(r io.Reader) (io.Reader, error) { return snappy.NewReader(r), nil },
	}
)

type Compressor interface {
	ContentType() string

	WrapWriter(io.Writer) (io.Writer, error)
	WrapReader(io.Reader) (io.Reader, error)
}

type StaticCompressor struct {
	Type       string
	WriterFunc func(io.Writer) (io.Writer, error)
	ReaderFunc func(io.Reader) (io.Reader, error)
}

func (sc *StaticCompressor) ContentType() string { return sc.Type }

func (sc *StaticCompressor) WrapWriter(w io.Writer) (io.Writer, error) {
	return sc.WriterFunc(w)
}

func (sc *StaticCompressor) WrapReader(r io.Reader) (io.Reader, error) {
	return sc.ReaderFunc(r)
}

func CombineCompressors(cs ...Compressor) Compressor {
	switch len(cs) {
	case 0:
		return NopCompressor
	case 1:
		return cs[0]
	}

	return MultiCompressor(cs)
}

var NopCompressor Compressor = nopCompressor{}

type nopCompressor struct{}

func (nopCompressor) ContentType() string { return "" }

func (nopCompressor) WrapWriter(w io.Writer) (io.Writer, error) { return w, nil }
func (nopCompressor) WrapReader(r io.Reader) (io.Reader, error) { return r, nil }

type MultiCompressor []Compressor

func (cs MultiCompressor) ContentType() string {
	var fs []string

	for _, c := range cs {
		fs = append(fs, c.ContentType())
	}

	return strings.Join(fs, "+")
}

func (cs MultiCompressor) WrapWriter(w io.Writer) (io.Writer, error) {
	var err error

	for _, c := range cs {
		w, err = c.WrapWriter(w)

		if err != nil {
			return w, err
		}
	}

	return w, nil
}
func (cs MultiCompressor) WrapReader(r io.Reader) (io.Reader, error) {
	var err error

	for _, c := range cs {
		r, err = c.WrapReader(r)

		if err != nil {
			return r, err
		}
	}

	return r, nil
}
