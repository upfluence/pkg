package roundrobin

import (
	"github.com/upfluence/sql"
	"github.com/upfluence/sql/backend/balancer"
)

// NewDB: this package is DEPRECATED use the backend/balancer package.
// This package will be removed in the next major release
func NewDB(dbs ...sql.DB) sql.DB {
	return balancer.NewDB(balancer.RoundRobinBalancerBuilder, dbs...)
}
