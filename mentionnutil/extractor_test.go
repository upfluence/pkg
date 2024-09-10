package mentionnutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractMentions(t *testing.T) {
	for _, tt := range []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name: "empty",
		},
		{
			name:  "no mentions",
			input: "foo bar",
		},
		{
			name:   "multi mentions",
			input:  "foo bar @buz. @taz_ @dot.name hello @buz @with_1",
			expect: []string{"buz", "taz_", "dot.name", "with_1"},
		},
		{
			name:   "youtube url",
			input:  "foo bar @buz. https://www.youtube.com/@SalonViking test",
			expect: []string{"buz", "SalonViking"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatch(t, tt.expect, ExtractMentions(tt.input))
		})
	}
}
