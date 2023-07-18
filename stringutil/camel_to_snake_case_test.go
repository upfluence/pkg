package stringutil

import (
	"testing"
)

func TestCamelToSnakeCase(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out string
	}{
		{"Camel", "camel"},
		{"VeryVeryLongCamelCase", "very_very_long_camel_case"},
		{"WithAcronymLikeURL", "with_acronym_like_url"},
	} {
		if out := CamelToSnakeCase(tt.in); tt.out != out {
			t.Errorf("CamelToSnakeCase(%q) = %q wanted: %q", tt.in, out, tt.out)
		}
	}
}
