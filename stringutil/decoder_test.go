package stringutil

import (
	"fmt"
	"testing"
)

func TestDecodeToUTF8(t *testing.T) {
	for _, testcase := range []struct {
		Value    string
		Expected string
	}{
		{"test", "test"},
		{"testé", "testé"},
		{"test\xc3", "testÃ"},
		{"\xc3", "Ã"},
	} {
		if testcase.Expected != DecodeToUTF8(testcase.Value) {
			t.Errorf(
				fmt.Sprintf(
					"Fail to decode string '%x'",
					testcase.Value,
				),
			)

		}
	}
}
