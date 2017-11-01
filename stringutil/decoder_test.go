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
