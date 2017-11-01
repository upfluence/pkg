package stringutil

import "golang.org/x/text/encoding/charmap"

func DecodeToUTF8(s string) (string, error) {
	return charmap.ISO8859_1.NewDecoder().String(s)
}
