package balancer

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/upfluence/sql"
)

var (
	RoundRobinBalancerBuilder   = BalancerBuilderFunc(buildRoundRobinBalancer)
	LeastPendingBalancerBuilder = BalancerBuilderFunc(buildLeastPendingBalancer)

	errNoDB = errors.New("backend/balancer: No DB registered")
)

type CloseFunc func(error)

type BalancerBuilder interface {
	Build([]sql.DB) Balancer
}

type BalancerBuilderFunc func([]sql.DB) Balancer

func (fn BalancerBuilderFunc) Build(dbs []sql.DB) Balancer { return fn(dbs) }

type Balancer interface {
	Get(context.Context) (sql.DB, CloseFunc, error)
}

func buildRoundRobinBalancer(dbs []sql.DB) Balancer {
	return &roundRobinBalancer{dbs: dbs}
}

type roundRobinBalancer struct {
	dbs []sql.DB

	mu sync.Mutex
	i  int
}

func (rrb *roundRobinBalancer) Get(context.Context) (sql.DB, CloseFunc, error) {
	rrb.mu.Lock()
	defer rrb.mu.Unlock()

	if len(rrb.dbs) == 0 {
		return nil, nil, errNoDB
	}

	db := rrb.dbs[rrb.i]
	rrb.i = (rrb.i + 1) % len(rrb.dbs)

	return db, func(error) {}, nil
}

func buildLeastPendingBalancer(dbs []sql.DB) Balancer {
	return &leastPendingBalancer{
		dbs:      dbs,
		pendings: make(map[sql.DB]int, len(dbs)),
	}
}

type leastPendingBalancer struct {
	mu sync.Mutex

	dbs      []sql.DB
	pendings map[sql.DB]int
}

func (lpb *leastPendingBalancer) Len() int { return len(lpb.dbs) }

func (lpb *leastPendingBalancer) Less(i, j int) bool {
	return lpb.pendings[lpb.dbs[i]] < lpb.pendings[lpb.dbs[j]]
}

func (lpb *leastPendingBalancer) Swap(i, j int) {
	lpb.dbs[i], lpb.dbs[j] = lpb.dbs[j], lpb.dbs[i]
}

func (lpb *leastPendingBalancer) Get(context.Context) (sql.DB, CloseFunc, error) {
	lpb.mu.Lock()
	defer lpb.mu.Unlock()

	if len(lpb.dbs) == 0 {
		return nil, nil, errNoDB
	}

	db := lpb.dbs[0]

	lpb.pendings[db]++
	sort.Sort(lpb)

	return db, func(error) {
		lpb.mu.Lock()

		lpb.pendings[db]--
		sort.Sort(lpb)

		lpb.mu.Unlock()
	}, nil
}
