package stringutil

import "testing"

func TestDecodeToUTF8(t *testing.T) {
	for _, tt := range []struct {
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
		if out := DecodeToUTF8(tt.in); tt.out != out {
			t.Errorf("DecodeToTUF8(%q) = %q wanted: %q", tt.in, out, tt.out)
		}
	}
}

func TestDecodeToASCII(t *testing.T) {
	for _, tt := range []struct {
		in, out string
	}{
		{"étesté", "eteste"},
		{"RhôöÔÖne", "RhooOOne"},
	} {
		if out := DecodeToASCII(tt.in); tt.out != out {
			t.Errorf("DecodeToASCII(%q) = %q wanted: %q", tt.in, out, tt.out)
		}
	}
}
