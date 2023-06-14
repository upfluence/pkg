package ioutil

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitedWriter(t *testing.T) {
	for _, tt := range []struct {
		in string

		wantN          int
		wantOverflowed bool
		wantOut        string
	}{
		{},
		{
			in:      "f",
			wantN:   1,
			wantOut: "f",
		},
		{
			in:             "foof",
			wantN:          4,
			wantOut:        "foof",
			wantOverflowed: true,
		},
		{
			in:             "foobar",
			wantN:          6,
			wantOut:        "foob",
			wantOverflowed: true,
		},
	} {
		var (
			buf bytes.Buffer

			lw = LimitedWriter{Writer: &buf, N: 4}
		)

		n, err := io.WriteString(&lw, tt.in)

		assert.NoError(t, err)
		assert.Equal(t, n, tt.wantN)
		assert.Equal(t, lw.Overflowed(), tt.wantOverflowed)
		assert.Equal(t, buf.String(), tt.wantOut)
	}
}
