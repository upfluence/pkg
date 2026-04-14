// Package lru provides a generic intrusive circular doubly-linked list with
// an integrated node pool, shared by size and tinylfu policy implementations.
//
// The list uses an embedded sentinel node so that every pointer operation is
// branch-free. sentinel.Next is the front (LRU / oldest end); sentinel.Prev
// is the back (MRU / newest end).
//
// Node carries a second type parameter E for caller-defined metadata, stored
// inline with no heap allocation. The tinylfu policy sets E to *segment[K] to
// record each node's owning segment; the size policy sets E to struct{}.
package lru

import "sync"

// Node is an element of a [List]. Callers must not copy a Node after first
// use.
//
// K is the key type; E is an arbitrary extra payload stored inline.
type Node[K comparable, E any] struct {
	Key        K
	Prev, Next *Node[K, E]

	// Extra is an inline slot for caller-defined metadata. It is never read
	// or written by List itself.
	Extra E
}

// List is a circular doubly-linked list with an embedded sentinel node and an
// integrated node pool.
//
// The pool recycles nodes released via [List.Free] so that [List.Alloc]
// returns them without a heap allocation. In steady state (when the list is
// at capacity and an evicted front node is reused in-place) neither Alloc nor
// Free is called at all.
//
// Use [NewList] to construct a usable List; the zero value is not usable.
type List[K comparable, E any] struct {
	sentinel Node[K, E]
	Len      int
	pool     sync.Pool
}

// NewList returns a pointer to a new List ready for use.
func NewList[K comparable, E any]() *List[K, E] {
	l := &List[K, E]{}
	l.sentinel.Next = &l.sentinel
	l.sentinel.Prev = &l.sentinel
	l.pool.New = func() any { return new(Node[K, E]) }
	return l
}

// Front returns the front (LRU / oldest) node.
func (l *List[K, E]) Front() *Node[K, E] { return l.sentinel.Next }

// Back returns the back (MRU / newest) node.
func (l *List[K, E]) Back() *Node[K, E] { return l.sentinel.Prev }

// Sentinel returns the sentinel node. Callers can compare against it to
// detect list boundaries.
func (l *List[K, E]) Sentinel() *Node[K, E] { return &l.sentinel }

// Remove unlinks n from the list. n must be a member of this list.
func (l *List[K, E]) Remove(n *Node[K, E]) {
	n.Prev.Next = n.Next
	n.Next.Prev = n.Prev
	l.Len--
}

// PushBack links n at the back (MRU end) of the list.
func (l *List[K, E]) PushBack(n *Node[K, E]) {
	s := &l.sentinel
	n.Prev = s.Prev
	n.Next = s
	s.Prev.Next = n
	s.Prev = n
	l.Len++
}

// MoveToBack moves n to the back (MRU end). It is a no-op if n is already
// there. Len is unchanged.
func (l *List[K, E]) MoveToBack(n *Node[K, E]) {
	if n == l.sentinel.Prev {
		return
	}
	// Unlink.
	n.Prev.Next = n.Next
	n.Next.Prev = n.Prev
	// Relink at back.
	s := &l.sentinel
	n.Prev = s.Prev
	n.Next = s
	s.Prev.Next = n
	s.Prev = n
}

// Alloc returns a recycled or freshly allocated Node ready for use.
func (l *List[K, E]) Alloc() *Node[K, E] {
	return l.pool.Get().(*Node[K, E])
}

// Free returns n to the pool for future reuse via Alloc.
// n must not be a member of any list when Free is called.
func (l *List[K, E]) Free(n *Node[K, E]) {
	l.pool.Put(n)
}
