package csv

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type record struct {
	Name string `csv:"name"`
	Age  int    `csv:"age"`
}

func TestReader(t *testing.T) {
	for _, tt := range []struct {
		input         string
		wantRecords   []record
		wantLeftovers []map[string]string
	}{
		{
			input: `age
1
2`,
			wantRecords:   []record{{Age: 1}, {Age: 2}},
			wantLeftovers: []map[string]string{nil, nil},
		},
		{
			input: `name,age
foo,1
bar,2`,
			wantRecords:   []record{{Name: "foo", Age: 1}, {Name: "bar", Age: 2}},
			wantLeftovers: []map[string]string{nil, nil},
		},
		{
			input: `name,age,extra
foo,1,aa
bar,2,bb`,
			wantRecords:   []record{{Name: "foo", Age: 1}, {Name: "bar", Age: 2}},
			wantLeftovers: []map[string]string{{"extra": "aa"}, {"extra": "bb"}},
		},
	} {
		var (
			gotRecords   []record
			gotLeftovers []map[string]string

			r, err = NewReader(strings.NewReader(tt.input))
		)

		require.NoError(t, err)

		for {
			var rec record

			leftover, err := r.Read(&rec)

			if err == io.EOF {
				break
			}

			require.NoError(t, err)

			gotRecords = append(gotRecords, rec)
			gotLeftovers = append(gotLeftovers, leftover)
		}

		assert.Equal(t, tt.wantRecords, gotRecords)
		assert.Equal(t, tt.wantLeftovers, gotLeftovers)
	}
}
