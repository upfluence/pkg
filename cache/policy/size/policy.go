package size

import (
	"github.com/upfluence/pkg/cache/v2/policy/size"
)

type Policy = size.Policy

func NewLRUPolicy(sz int) *Policy {
	return size.NewLRUPolicy(sz)
}
