package thriftutil

import (
	"testing"
)

func TestSanitizeCamelCaseThriftEnumValue(t *testing.T) {
	for _, tt := range []struct {
		in  string
		out string
	}{
		{"ContentType_InstagramMedia", "instagram_media"},
		{"ContentType_Unknown", "unknown"},
		{"regular string", "regular string"},
		{"MFAVector_RECOVERY_CODE", "recovery_code"},
	} {
		if out := SanitizeCamelCaseThriftEnumValue(tt.in); tt.out != out {
			t.Errorf("SanitizeCamelCaseThriftEnumValue(%q) = %q wanted: %q", tt.in, out, tt.out)
		}
	}
}
