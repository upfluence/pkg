package stringutil

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/upfluence/pkg/log"
)

var cmap = charmap.ISO8859_1
var defaultDecoder = charmap.ISO8859_1.NewDecoder()

type ASCIIDecodeOption func(*asciiDecodeOptions)

type asciiDecodeOptions struct {
	composer   transform.Transformer
	decomposer transform.Transformer
}

func NKFD(opts *asciiDecodeOptions) {
	opts.composer = norm.NFKC
	opts.decomposer = norm.NFKD
}

var defaultDecodeOptions = asciiDecodeOptions{
	composer:   norm.NFC,
	decomposer: norm.NFD,
}

func DecodeToUTF8(s string) string {
	s = strings.Replace(s, "\x00", "", -1)

	if IsUTF8(s) {
		return s
	}

	s, err := defaultDecoder.String(s)

	if err != nil {
		log.Warning(err.Error())

		return ""
	}

	return s
}

func IsUTF8(s string) bool {
	return utf8.ValidString(s)
}

func isAboveAscii(r rune) bool {
	return r > unicode.MaxASCII
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func IsASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}

	return true
}

func DecodeToASCII(s string, opts ...ASCIIDecodeOption) string {
	if IsASCII(s) {
		return s
	}

	os := defaultDecodeOptions

	for _, opt := range opts {
		opt(&os)
	}

	var (
		t = transform.Chain(os.decomposer, transform.RemoveFunc(isMn), transform.RemoveFunc(isAboveAscii), os.composer)

		result, _, err = transform.String(t, s)
	)

	if err != nil {
		log.Warning(err)
		return ""
	}

	return result
}
