package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluralize(t *testing.T) {
	for _, tt := range []struct {
		singular string
		plural   string
	}{
		{singular: "try", plural: "tries"},
		{singular: "book", plural: "books"},
		{singular: "medium", plural: "media"},
	} {
		assert.Equal(t, tt.plural, DefaultInflector.Pluralize(tt.singular))
		assert.Equal(t, tt.singular, DefaultInflector.Singularize(tt.plural))
	}
}
