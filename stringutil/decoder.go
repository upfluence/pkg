package stringutil

import (
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

func DecodeToUTF8(s string) string {
	if IsUTF8(s) {
		return s
	}

	s, err := charmap.ISO8859_1.NewDecoder().String(s)

	if err != nil {
		return ""
	}

	return s
}

func IsUTF8(s string) bool {
	return utf8.ValidString(s)
}
