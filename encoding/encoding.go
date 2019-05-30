package encoding

import (
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"

	"github.com/golang/snappy"
)

var (
	GZipEncoding = &StaticEncoding{
		Type:       "gzip",
		WriterFunc: func(w io.Writer) (io.Writer, error) { return gzip.NewWriter(w), nil },
		ReaderFunc: func(r io.Reader) (io.Reader, error) { return gzip.NewReader(r) },
	}

	SnappyEncoding = &StaticEncoding{
		Type:       "snappy",
		WriterFunc: func(w io.Writer) (io.Writer, error) { return snappy.NewWriter(w), nil },
		ReaderFunc: func(r io.Reader) (io.Reader, error) { return snappy.NewReader(r), nil },
	}

	Base64Encoding = &StaticEncoding{
		Type: "base64",
		WriterFunc: func(w io.Writer) (io.Writer, error) {
			return base64.NewEncoder(base64.StdEncoding, w), nil
		},
		ReaderFunc: func(r io.Reader) (io.Reader, error) {
			return base64.NewDecoder(base64.StdEncoding, r), nil
		},
	}

	NopEncoderFunc = func(w io.Writer) (io.Writer, error) { return w, nil }
	NopDecoderFunc = func(r io.Reader) (io.Reader, error) { return r, nil }
)

type EncoderFunc func(io.Writer) (io.Writer, error)
type DecoderFunc func(io.Reader) (io.Reader, error)

type Encoding interface {
	ContentType() string

	WrapWriter(io.Writer) (io.Writer, error)
	WrapReader(io.Reader) (io.Reader, error)
}

type StaticEncoding struct {
	Type       string
	WriterFunc EncoderFunc
	ReaderFunc DecoderFunc
}

func (sc *StaticEncoding) ContentType() string { return sc.Type }

func (sc *StaticEncoding) WrapWriter(w io.Writer) (io.Writer, error) {
	return sc.WriterFunc(w)
}

func (sc *StaticEncoding) WrapReader(r io.Reader) (io.Reader, error) {
	return sc.ReaderFunc(r)
}

func CombineEncodings(cs ...Encoding) Encoding {
	switch len(cs) {
	case 0:
		return NopEncoding
	case 1:
		return cs[0]
	}

	return MultiEncoding(cs)
}

var NopEncoding Encoding = nopEncoding{}

type nopEncoding struct{}

func (nopEncoding) ContentType() string { return "" }

func (nopEncoding) WrapWriter(w io.Writer) (io.Writer, error) { return w, nil }
func (nopEncoding) WrapReader(r io.Reader) (io.Reader, error) { return r, nil }

type MultiEncoding []Encoding

func (cs MultiEncoding) ContentType() string {
	var fs []string

	for _, c := range cs {
		fs = append(fs, c.ContentType())
	}

	return strings.Join(fs, "+")
}

func (cs MultiEncoding) WrapWriter(w io.Writer) (io.Writer, error) {
	var err error

	for _, c := range cs {
		w, err = c.WrapWriter(w)

		if err != nil {
			return w, err
		}
	}

	return w, nil
}
func (cs MultiEncoding) WrapReader(r io.Reader) (io.Reader, error) {
	var err error

	for _, c := range cs {
		r, err = c.WrapReader(r)

		if err != nil {
			return r, err
		}
	}

	return r, nil
}
