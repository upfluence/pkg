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

var defaultDecoder = charmap.ISO8859_1.NewDecoder()

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

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func DecodeToASCII(s string) string {
	var (
		b = make([]byte, len(s))
		t = transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)

		nDst, _, err = t.Transform(b, []byte(s), true)
	)

	if err != nil {
		log.Warning(err)
		return ""
	}

	return string(b[:nDst])
}
