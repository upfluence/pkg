package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	var m = map[string]int{
		"foo":  2,
		"fizz": 1,
	}

	assert.ElementsMatch(t, []string{"foo", "fizz"}, Keys(m))
}

func TestValues(t *testing.T) {
	var m = map[string]int{
		"foo":  2,
		"fizz": 1,
	}

	assert.ElementsMatch(t, []int{1, 2}, Values(m))
}
