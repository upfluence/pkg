package int64util

import "sort"

type Slice []int64

func (s Slice) Len() int           { return len(s) }
func (s Slice) Less(i, j int) bool { return s[i] < s[j] }
func (s Slice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s Slice) Sort()              { sort.Sort(s) }
func (s Slice) Search(v int64) int { return Search(s, v) }

func (s Slice) ToSet() *Set {
	var res Set

	res.Add(s...)

	return &res
}

// Deprecated: Consider using slices.Sort, slices.Index and upfluence/pkg/slices functions.
func ToSlice(int64s []int64) Slice {
	var s = Slice(int64s)

	sort.Sort(s)

	return s
}

func Search(s Slice, v int64) int {
	return sort.Search(len(s), func(i int) bool { return s[i] >= v })
}
