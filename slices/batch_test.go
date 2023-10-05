package slices

import (
	"reflect"
	"testing"
)

func TestBatch(t *testing.T) {
	for _, tt := range []struct {
		slice []string
		size  int
		want  [][]string
	}{
		{size: 5},
		{slice: []string{}, size: 5},
		{slice: []string{"foo"}, size: 2, want: [][]string{{"foo"}}},
		{slice: []string{"foo", "bar"}, size: 2, want: [][]string{{"foo", "bar"}}},
		{
			slice: []string{"foo", "bar", "biz"},
			size:  2,
			want:  [][]string{{"foo", "bar"}, {"biz"}},
		},
	} {
		if out := Batch(tt.slice, tt.size); !reflect.DeepEqual(tt.want, out) {
			t.Errorf(
				"Batch(%v, %d) = %+v [ want: %+v ]",
				tt.slice,
				tt.size,
				out,
				tt.want,
			)
		}
	}
}
