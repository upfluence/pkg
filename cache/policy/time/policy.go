package time

import (
	"time"

	tpol "github.com/upfluence/pkg/cache/v2/policy/time"
)

type Policy = tpol.Policy[string]

func NewIdlePolicy(ttl time.Duration) *Policy {
	return tpol.NewIdlePolicy[string](ttl)
}

func NewLifetimePolicy(ttl time.Duration) *Policy {
	return tpol.NewLifetimePolicy[string](ttl)
}
