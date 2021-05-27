package int64util

import (
	"reflect"
	"testing"
)

func TestBatch(t *testing.T) {
	for _, tt := range []struct {
		slice []int64
		size  int
		want  [][]int64
	}{
		{size: 5},
		{slice: []int64{}, size: 5},
		{slice: []int64{123}, size: 2, want: [][]int64{{123}}},
		{slice: []int64{123, 456}, size: 2, want: [][]int64{{123, 456}}},
		{
			slice: []int64{123, 456, 789},
			size:  2,
			want:  [][]int64{{123, 456}, {789}},
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
