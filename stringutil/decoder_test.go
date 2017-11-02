package stringutil

import (
	"fmt"
	"testing"
)

func TestDecodeToUTF8(t *testing.T) {
	for _, testcase := range []struct {
		in, out string
	}{
		{"test", "test"},
		{"testé", "testé"},
		{"test\xc3", "testÃ"},
		{"\x81\x9F", "\xC2\x81\xC2\x9F"},
		{"\x00", ""},
		{"test \x00 test", "test  test"},
		{"\xc3", "Ã"},
	} {
		if testcase.out != DecodeToUTF8(testcase.in) {
			t.Errorf(
				fmt.Sprintf(
					"Fail to decode string '%s'",
					testcase.in,
				),
			)

		}
	}
}
