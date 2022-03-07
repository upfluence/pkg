package stringutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullablePtr(t *testing.T) {
	var str = "string"

	for _, tt := range []struct {
		s    string
		want *string
	}{
		{},
		{s: str, want: &str},
	} {
		t.Run(fmt.Sprint("ptr value of ", tt.s), func(t *testing.T) {
			assert.Equal(t, tt.want, NullablePtr(tt.s))
		})
	}
}
