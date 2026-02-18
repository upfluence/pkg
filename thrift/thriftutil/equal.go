package thriftutil

import (
	"github.com/upfluence/pkg/v2/thrift/hasher"
	"github.com/upfluence/thrift/lib/go/thrift"
)

var defaultHasher = hasher.NewFNVHasher64()

func Equal[T any, PT interface {
	thrift.TStruct
	*T
}](a, b PT) bool {
	ha, err := defaultHasher.Hash(a)

	if err != nil {
		return false
	}

	hb, err := defaultHasher.Hash(b)

	if err != nil {
		return false
	}

	return ha == hb
}
