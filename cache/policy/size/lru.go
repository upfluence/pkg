package size

import (
	"github.com/upfluence/pkg/v2/cache/policy"
	"github.com/upfluence/pkg/v2/cache/policy/internal/lru"
)

// lruBackend implements backend using a doubly-linked LRU list.
type lruBackend[K comparable] struct {
	l    *lru.List[K, struct{}]
	size int
}

func (b *lruBackend[K]) insert(k K) (K, bool, *lru.Node[K, struct{}]) {
	var (
		n       *lru.Node[K, struct{}]
		evicted K
		ok      bool
	)

	if b.l.Len >= b.size {
		n = b.l.Front()
		evicted = n.Key
		b.l.Remove(n)
		ok = true
	} else {
		n = b.l.Alloc()
	}

	n.Key = k
	b.l.PushBack(n)

	return evicted, ok, n
}

func (b *lruBackend[K]) remove(n *lru.Node[K, struct{}]) {
	b.l.Remove(n)
	b.l.Free(n)
}

func (b *lruBackend[K]) get(_ K, n *lru.Node[K, struct{}]) {
	b.l.MoveToBack(n)
}

// NewLRUPolicy returns an LRU eviction policy capped at size entries.
func NewLRUPolicy[K comparable](size int) policy.EvictionPolicy[K] {
	return newPolicy[K, *lru.Node[K, struct{}]](
		&lruBackend[K]{
			size: size,
			l:    lru.NewList[K, struct{}](),
		},
		size,
	)
}
