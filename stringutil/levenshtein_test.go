package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevenshteinDistanceCalculator(t *testing.T) {
	for _, tt := range []struct {
		name string

		indel int
		sub   int
		s1    string
		s2    string
		want  int
	}{
		{
			name: "equal strings",

			indel: 1,
			sub:   1,
			s1:    "test",
			s2:    "test",
			want:  0,
		},
		{
			name: "completely different strings",

			indel: 1,
			sub:   1,
			s1:    "abcdfg",
			s2:    "test",
			want:  6,
		},
		{
			name: "one letter difference",

			indel: 1,
			sub:   2,
			s1:    "tsst",
			s2:    "test",
			want:  2,
		},
		{
			name: "one extra letter difference",

			indel: 3,
			sub:   1,
			s1:    "test",
			s2:    "testt",
			want:  3,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ld := NewLevenshteinDistanceCalculator(tt.indel, tt.sub)

			res := ld.LevenshteinDist(tt.s1, tt.s2)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestLevenshteinDist(t *testing.T) {
	for _, tt := range []struct {
		name string

		s1   string
		s2   string
		want float64
	}{
		{
			name: "equal strings",

			s1:   "test",
			s2:   "test",
			want: 0,
		},
		{
			name: "completely different strings",

			s1:   "abc",
			s2:   "defgh",
			want: 1,
		},
		{
			name: "similar words",

			s1:   "tss",
			s2:   "test",
			want: 0.5,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			res := LevenshteinDist(tt.s1, tt.s2)
			assert.InDelta(t, tt.want, res, 0.01)
		})
	}
}
