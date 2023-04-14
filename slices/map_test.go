//go:build go1.20

package slices

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/upfluence/errors"
)

func TestMap(t *testing.T) {
	var merr = errors.New("test")

	for _, tt := range []struct {
		name    string
		in      []string
		fn      func(string) (int64, error)
		want    []int64
		wantErr error
	}{
		{
			name: "error",
			in:   []string{"fail"},
			fn: func(string) (int64, error) {
				return 0, merr
			},
			wantErr: merr,
		},
		{
			name: "success",
			in:   []string{"1", "22"},
			fn: func(s string) (int64, error) {
				return strconv.ParseInt(s, 10, 64)
			},
			want: []int64{1, 22},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapSlice(tt.in, tt.fn)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
