package timeutil

import (
	"time"

	"github.com/upfluence/pkg/pointers"
)

// UnixOrrNil returns the unix timestamp or nil if time is zero.
func UnixOrNil(t time.Time) *int64 {
	if t.IsZero() {
		return nil
	}

	return pointers.Ptr(t.Unix())
}

func UnixOrZero(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}

	return t.Unix()
}
