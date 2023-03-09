package stringutil

import "unicode/utf8"

// DistanceCalculator is the levenshtein distance calculator interface.
type DistanceCalculator interface {
	// LevenshteinDist calculates levenshtein distance between two utf-8 encoded strings
	LevenshteinDist(string, string) int
}

// NewLevenshteinDistanceCalculator creates a new levenshtein distance calculator where indel is increment/deletion cost
// and sub is the substitution cost.
func NewLevenshteinDistanceCalculator(indel, sub int) DistanceCalculator {
	return &calculator{indel, sub}
}

type calculator struct {
	indel, sub int
}

// https://en.wikibooks.org/wiki/Algorithm_Implementation/Strings/Levenshtein_distance#C
func (c *calculator) LevenshteinDist(s1, s2 string) int {
	l := utf8.RuneCountInString(s1)
	m := make([]int, l+1)

	for i := 1; i <= l; i++ {
		m[i] = i * c.indel
	}

	var y int

	lastdiag, x := 0, 1
	for _, rx := range s2 {
		m[0], lastdiag, y = x*c.indel, (x-1)*c.indel, 1
		for _, ry := range s1 {
			m[y], lastdiag = min3(m[y]+c.indel, m[y-1]+c.indel, lastdiag+c.subCost(rx, ry)), m[y]
			y++
		}
		x++
	}

	return m[l]
}

func (c *calculator) subCost(r1, r2 rune) int {
	if r1 == r2 {
		return 0
	}

	return c.sub
}

func min3(a, b, c int) int {
	return min(a, min(b, c))
}

func max(a, b int) int {
	if a < b {
		return b
	}

	return a
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

var defaultCalculator = NewLevenshteinDistanceCalculator(1, 1)

// LevenshteinDist is a convenience function for a levenshtein distance calculator with equal costs.
func LevenshteinDist(s1, s2 string) float64 {
	return float64(defaultCalculator.LevenshteinDist(s1, s2)) / float64(max(len(s1), len(s2)))
}
