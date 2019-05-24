package stringutil

import (
  "io/ioutil"
  "os"
  "testing"
)

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
    {"輸入罕見異體字或者輸入插圖", "輸入罕見異體字或者輸入插圖"},
	} {
		if out := DecodeToUTF8(tt.in); tt.out != out {
			t.Errorf("DecodeToTUF8(%q) = %q wanted: %q", tt.in, out, tt.out)
		}
	}
}

func TestDecodeToUTF8FromFile(t *testing.T) {
  for _, tt := range []struct {
    in, out string
  }{
    { "testdata/chinese.txt", "" },
    { "testdata/chinese_original.txt", "輸入罕見異體字或者輸入插圖" },
  } {
    f, err := os.Open(tt.in)

    if err != nil {
      t.Fatalf("Can't open file %q:%v", tt.in, err)
    }

    defer f.Close()
    buf, err := ioutil.ReadAll(f)

    if err != nil {
      t.Fatalf("Can't read buffer %q:%v", tt.in, err)
    }

		if out := DecodeToUTF8(string(buf)); tt.out != out {
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
		{"東京都, JP", ", JP"},
	} {
		if out := DecodeToASCII(tt.in); tt.out != out {
			t.Errorf("DecodeToASCII(%q) = %q wanted: %q", tt.in, out, tt.out)
		}
	}
}
