package policy

import (
	policy "github.com/upfluence/pkg/cache/v2/policy"
)

var ErrClosed = policy.ErrClosed

type EvictionPolicy = policy.EvictionPolicy[string]

func CombinePolicies(ps ...EvictionPolicy) EvictionPolicy {
	return policy.CombinePolicies(ps...)
}

type NopPolicy = policy.NopPolicy[string]

type OpType = policy.OpType

const (
	Get   = policy.Get
	Set   = policy.Set
	Evict = policy.Evict
)
