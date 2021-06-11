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

func DecodeToASCII(s string) string {
	if IsASCII(s) {
		return s
	}

	var (
		t = transform.Chain(norm.NFD, transform.RemoveFunc(isMn), transform.RemoveFunc(isAboveAscii), norm.NFC)

		result, _, err = transform.String(t, s)
	)

	if err != nil {
		log.Warning(err)
		return ""
	}

	return result
}
