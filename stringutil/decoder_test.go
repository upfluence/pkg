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
	var nfkd = []ASCIIDecodeOption{NKFD}

	for _, tt := range []struct {
		in, out string
		opts    []ASCIIDecodeOption
	}{
		{
			in:  "étesté",
			out: "eteste",
		},
		{
			in:  "RhôöÔÖne",
			out: "RhooOOne",
		},
		{
			in:  "東京都, JP",
			out: ", JP",
		},
		{
			in:  "Collaboration: 𝕸𝖎𝖆𝖒𝖎 🌞 x KiwiKurve",
			out: "Collaboration: x KiwiKurve",
		},
		{
			in:  "back soon ✌🏽📍ashleyrchand@gmail.com",
			out: "back soon ashleyrchand@gmail.com",
		},
		{
			in:   "Golden Girl 🌴\n🌿Discounts/links⬇️\nPR/Collab📧spfpleaseka𝐓ie@gmail.com",
			out:  "Golden Girl Discounts/links PR/Collab spfpleasekaTie@gmail.com",
			opts: nfkd,
		},
		{
			in:  "foo",
			out: "foo",
		},
		{
			in:   "foo",
			out:  "foo",
			opts: nfkd,
		},
		{
			in:   "𝐓𝐲𝐫𝐚𝐉𝐚𝐧𝐞𝐚✨🎨",
			out:  "TyraJanea",
			opts: nfkd,
		},
	} {
		if out := DecodeToASCII(tt.in, tt.opts...); tt.out != out {
			t.Errorf("DecodeToASCII(%q) = %q wanted: %q", tt.in, out, tt.out)
		}
	}
}
