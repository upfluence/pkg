package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReduceWith(t *testing.T) {
	for _, tt := range []struct {
		slice []int
		acc   int
		fn    func(int, int) int
		want  int
	}{
		{slice: []int{1, 2, 3}, acc: 0, fn: func(a, b int) int { return a + b }, want: 6},
		{slice: []int{1, 2, 3}, acc: 1, fn: func(a, b int) int { return a * b }, want: 6},
		{slice: []int{}, acc: 0, fn: func(a, b int) int { return a + b }, want: 0},
		{slice: []int{1}, acc: 0, fn: func(a, b int) int { return a + b }, want: 1},
	} {
		t.Run("", func(t *testing.T) {
			got := ReduceFrom(tt.slice, tt.acc, tt.fn)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReduce(t *testing.T) {
	for _, tt := range []struct {
		slice []int
		fn    func(int, int) int
		want  int
	}{
		{slice: []int{1, 2, 3}, fn: func(a, b int) int { return a + b }, want: 6},
		{slice: []int{1, 2, 3}, fn: func(a, b int) int { return a * b }, want: 0},
		{slice: []int{}, fn: func(a, b int) int { return a + b }, want: 0},
		{slice: []int{1}, fn: func(a, b int) int { return a + b }, want: 1},
	} {
		got := Reduce(tt.slice, tt.fn)
		assert.Equal(t, tt.want, got)
	}
}

func TestSum(t *testing.T) {
	t.Run("integer", func(t *testing.T) {
		assert.Equal(t, 3, Reduce([]int{1, 2}, Sum))
		assert.Equal(t, 25, ReduceFrom([]int{1, 2}, 22, Sum))
	})
	t.Run("float", func(t *testing.T) {
		assert.Equal(t, 3.4, Reduce([]float64{1.1, 2.3}, Sum))
		assert.Equal(t, 25.52, ReduceFrom([]float64{1.1, 2.2}, 22.22, Sum))
	})
	t.Run("string", func(t *testing.T) {
		assert.Equal(t, "foobar", Reduce([]string{"foo", "bar"}, Sum))
		assert.Equal(t, "buzfoobar", ReduceFrom([]string{"foo", "bar"}, "buz", Sum))
	})
}
