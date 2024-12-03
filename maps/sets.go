package maps

import (
	"sync"

	"golang.org/x/exp/maps"
)

// Set is a concurrency safe wrapper around a map to ensure uniqueness through
// a collection.
// While the underlying map can be manually manipulated, read and
// mutation should be done with through methods to ensure
// concurrency safety.
type Set[T comparable] struct {
	mu sync.RWMutex

	Set map[T]struct{}
}

func (s *Set[T]) Add(vs ...T) {
	if len(vs) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Set == nil {
		s.Set = make(map[T]struct{}, len(vs))
	}

	for _, v := range vs {
		s.Set[v] = struct{}{}
	}
}

func (s *Set[T]) Has(v T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Set) == 0 {
		return false
	}

	_, ok := s.Set[v]

	return ok
}

func (s *Set[T]) Keys() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Set) == 0 {
		return nil
	}

	return Keys(s.Set)
}

func (s *Set[T]) Delete(vs ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range vs {
		delete(s.Set, v)
	}
}

// IndexedSet is a Set where the key of the map is extracted from
// Fn.
// Adding values with the same key will replace the previous value
type IndexedSet[K comparable, T any] struct {
	mu sync.RWMutex

	Fn  func(T) K
	Set map[K]T
}

func (s *IndexedSet[K, T]) Add(vs ...T) {
	if len(vs) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Set == nil {
		s.Set = make(map[K]T, len(vs))
	}

	for _, v := range vs {
		s.Set[s.Fn(v)] = v
	}
}

func (s *IndexedSet[K, T]) Has(k K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Set) == 0 {
		return false
	}

	_, ok := s.Set[k]

	return ok
}

func (s *IndexedSet[K, T]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Set) == 0 {
		return nil
	}

	return Keys(s.Set)
}

func (s *IndexedSet[K, T]) Values() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Set) == 0 {
		return nil
	}

	return maps.Values(s.Set)
}

func (s *IndexedSet[K, T]) Delete(ks ...K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, k := range ks {
		delete(s.Set, k)
	}
}

// Keys is a dropin replacement for x/maps.Keys
func Keys[M ~map[K]T, K comparable, T any](m M) []K {
	return maps.Keys(m)
}

// Values is a dropin replacement for x/maps.Values
func Values[M ~map[K]T, K comparable, T any](m M) []T {
	return maps.Values(m)
}
