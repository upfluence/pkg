package int64util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullablePtr(t *testing.T) {
	var id int64 = 1

	for _, tt := range []struct {
		i    int64
		want *int64
	}{
		{},
		{i: id, want: &id},
	} {
		t.Run(fmt.Sprint("ptr value of ", tt.i), func(t *testing.T) {
			assert.Equal(t, tt.want, NullablePtr(tt.i))
		})
	}
}
